package http

import (
	"context"
	gotls "crypto/tls"
	"net/http"
	"strings"
	"time"

	goreality "github.com/xtls/reality"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	http_proto "github.com/xtls/xray-core/common/protocol/http"
	"github.com/xtls/xray-core/common/signal/done"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/reality"
	"github.com/xtls/xray-core/transport/internet/tls"
)

type Listener struct {
	server  *http.Server
	handler internet.ConnHandler
	local   net.Addr
	config  *Config
}

func (l *Listener) Addr() net.Addr {
	return l.local
}

func (l *Listener) Close() error {
	return l.server.Close()
}

func (l *Listener) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	host := request.Host
	l.config.applyHeader(writer.Header())
	if l.config.Host != "" && l.config.Host != host {
		writer.WriteHeader(404)
		return
	}
	path := l.config.getNormalizedPath()
	if !strings.HasPrefix(request.URL.Path, path) {
		writer.WriteHeader(404)
		return
	}

	writer.Header().Set("Cache-Control", "no-store")

	controler := http.NewResponseController(writer)
	controler.EnableFullDuplex()
	writer.WriteHeader(200)
	controler.Flush()

	remoteAddr := l.Addr()
	dest, err := net.ParseDestination(request.RemoteAddr)
	if err != nil {
		errors.LogInfoInner(context.Background(), err, "failed to parse request remote addr: ", request.RemoteAddr)
	} else {
		remoteAddr = &net.TCPAddr{
			IP:   dest.Address.IP(),
			Port: int(dest.Port),
		}
	}

	forwardedAddress := http_proto.ParseXForwardedFor(request.Header)
	if len(forwardedAddress) > 0 && forwardedAddress[0].Family().IsIP() {
		remoteAddr = &net.TCPAddr{
			IP:   forwardedAddress[0].IP(),
			Port: 0,
		}
	}

	fwriter := flushWriter{
		writer: writer,
		closed: done.New(),
		flush:  controler.Flush,
	}

	done := done.New()
	conn := &Connection{
		reader:        request.Body,
		writer:        fwriter,
		done:          done,
		local:         l.Addr(),
		remote:        remoteAddr,
		readDeadline:  controler.SetReadDeadline,
		writeDeadline: controler.SetWriteDeadline,
	}
	l.handler(conn)
	<-done.Wait()
}

func Listen(ctx context.Context, address net.Address, port net.Port, streamSettings *internet.MemoryStreamConfig, handler internet.ConnHandler) (internet.Listener, error) {
	httpSettings := streamSettings.ProtocolSettings.(*Config)
	tlsConfig := tls.ConfigFromStreamSettings(streamSettings)
	realityConfig := reality.ConfigFromStreamSettings(streamSettings)
	listener := &Listener{
		handler: handler,
		config:  httpSettings,
	}
	if port == net.Port(0) { // unix
		listener.local = &net.UnixAddr{
			Name: address.Domain(),
			Net:  "unix",
		}
	} else {
		listener.local = &net.TCPAddr{
			IP:   address.IP(),
			Port: int(port),
		}
	}

	if streamSettings.SocketSettings != nil && streamSettings.SocketSettings.AcceptProxyProtocol {
		errors.LogWarning(ctx, "accepting PROXY protocol")
	}

	server := &http.Server{
		Handler:           listener,
		ReadHeaderTimeout: time.Second * 4,
		Protocols:         &http.Protocols{},
	}
	server.Protocols.SetUnencryptedHTTP2(true)
	// just in case
	server.Protocols.SetHTTP1(true)

	listener.server = server
	go func() {
		var streamListener net.Listener
		var err error
		streamListener, err = internet.ListenSystem(ctx, listener.local, streamSettings.SocketSettings)
		if err != nil {
			errors.LogErrorInner(ctx, err, "failed to listen on ", address)
			return
		}

		if tlsConfig != nil {
			streamListener = gotls.NewListener(streamListener, tlsConfig.GetTLSConfig())
		} else if realityConfig != nil {
			streamListener = goreality.NewListener(streamListener, realityConfig.GetREALITYConfig())
		}
		err = server.Serve(streamListener)
		if err != nil {
			errors.LogInfoInner(ctx, err, "stopping serving HTTP2")
		}
	}()

	return listener, nil
}

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
