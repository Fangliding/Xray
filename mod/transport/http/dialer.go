package http

import (
	"context"
	gotls "crypto/tls"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/signal/done"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/reality"
	"github.com/xtls/xray-core/transport/internet/stat"
	"github.com/xtls/xray-core/transport/internet/tls"
	"golang.org/x/net/http2"
)

type dialerConf struct {
	net.Destination
	*internet.MemoryStreamConfig
}

var (
	globalDialerMap    = make(map[dialerConf]*http.Client)
	globalDialerAccess sync.Mutex
)

func getHTTPClient(dest net.Destination, streamSettings *internet.MemoryStreamConfig) (*http.Client, error) {
	globalDialerAccess.Lock()
	defer globalDialerAccess.Unlock()

	if client, found := globalDialerMap[dialerConf{dest, streamSettings}]; found {
		return client, nil
	}

	httpSettings := streamSettings.ProtocolSettings.(*Config)
	tlsConfig := tls.ConfigFromStreamSettings(streamSettings)
	realityConfig := reality.ConfigFromStreamSettings(streamSettings)
	sockopt := streamSettings.SocketSettings

	transport := &http2.Transport{
		AllowHTTP:          true,
		IdleConnTimeout:    net.ConnIdleTimeout,
		ReadIdleTimeout:    net.ChromeH2KeepAlivePeriod,
		DisableCompression: true,
		DialTLSContext: func(ctx context.Context, network string, addr string, cfg *gotls.Config) (net.Conn, error) {
			pconn, err := internet.DialSystem(ctx, dest, sockopt)
			if err != nil {
				errors.LogErrorInner(ctx, err, "failed to dial to "+addr)
				return nil, err
			}
			if tlsConfig != nil {
				var cn tls.Interface
				if fingerprint := tls.GetFingerprint(tlsConfig.Fingerprint); fingerprint != nil {
					cn = tls.UClient(pconn, tlsConfig.GetTLSConfig(), fingerprint).(*tls.UConn)
				} else {
					cn = tls.Client(pconn, tlsConfig.GetTLSConfig()).(*tls.Conn)
				}
				if err := cn.HandshakeContext(ctx); err != nil {
					errors.LogErrorInner(ctx, err, "failed to dial to "+addr)
					return nil, err
				}
				return cn, nil
			} else if realityConfig != nil {
				return reality.UClient(pconn, realityConfig, ctx, dest)
			} else if streamSettings.SecurityType == "none" || streamSettings.SecurityType == "" {
				return pconn, nil
			}
			panic("invalid security settings")
		},
	}

	switch {
	case httpSettings.IdleTimeout > 0:
		transport.ReadIdleTimeout = time.Second * time.Duration(httpSettings.IdleTimeout)
	case httpSettings.IdleTimeout < 0:
		// Disable if negative.
		transport.ReadIdleTimeout = 0
	}
	if httpSettings.HealthCheckTimeout > 0 {
		transport.PingTimeout = time.Second * time.Duration(httpSettings.HealthCheckTimeout)
	}

	client := &http.Client{
		Transport: transport,
	}
	globalDialerMap[dialerConf{dest, streamSettings}] = client
	return client, nil
}

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (stat.Connection, error) {
	httpSettings := streamSettings.ProtocolSettings.(*Config)
	client, err := getHTTPClient(dest, streamSettings)
	if err != nil {
		return nil, err
	}

	httpMethod := "POST"
	if httpSettings.Method != "" {
		httpMethod = httpSettings.Method
	}
	Host := httpSettings.Host
	if Host == "" {
		Host = dest.Address.String()
	}
	URL := &url.URL{
		Host: Host,
		Path: httpSettings.getNormalizedPath(),
	}
	if streamSettings.SecurityType == "none" || streamSettings.SecurityType == "" {
		URL.Scheme = "http"
	} else {
		URL.Scheme = "https"
	}

	preader, pwriter := io.Pipe()
	request, err := http.NewRequestWithContext(ctx, httpMethod, URL.String(), preader)
	if err != nil {
		return nil, err
	}
	httpSettings.applyHeader(request.Header)
	if len(request.Header.Values("Content-Type")) == 0 {
		request.Header.Set("Content-Type", "application/grpc")
	}

	wrc := &waitReadCloser{ready: done.New()}
	go func() {
		response, err := client.Do(request)
		if err != nil || response.StatusCode != 200 {
			if err != nil {
				errors.LogWarningInner(ctx, err, "failed to dial to ", dest)
			} else {
				errors.LogWarning(ctx, "unexpected status ", response.StatusCode)
			}
			wrc.Close()
			// Seems this dropping is unnecessary since we have healthy check
			// ———— ℱ

			// Abandon `client` if `client.Do(request)` failed
			// See https://github.com/golang/go/issues/30702
			// globalDialerAccess.Lock()
			// if globalDialerMap[dialerConf{dest, streamSettings}] == client {
			// 	delete(globalDialerMap, dialerConf{dest, streamSettings})
			// }
			// globalDialerAccess.Unlock()
			return
		}
		wrc.Set(response.Body)
	}()

	conn := &Connection{
		reader: wrc,
		writer: pwriter,
		done:   done.New(),
		remote: dest.RawNetAddr(),
	}

	return conn, nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}
