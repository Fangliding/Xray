package quic

import (
	"syscall"
	"time"

	"github.com/apernet/quic-go"
	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
)

type sysConn struct {
	conn *net.UDPConn
}

func wrapSysConn(rawConn *net.UDPConn, config *Config) (*sysConn, error) {
	return &sysConn{
		conn: rawConn,
		// auth: auth,
	}, nil
}

func (c *sysConn) ReadFrom(p []byte) (int, net.Addr, error) {
	return c.conn.ReadFrom(p)
}

func (c *sysConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	return c.conn.WriteTo(p, addr)
}

func (c *sysConn) Close() error {
	return c.conn.Close()
}

func (c *sysConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *sysConn) SetReadBuffer(bytes int) error {
	return c.conn.SetReadBuffer(bytes)
}

func (c *sysConn) SetWriteBuffer(bytes int) error {
	return c.conn.SetWriteBuffer(bytes)
}

func (c *sysConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *sysConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *sysConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *sysConn) SyscallConn() (syscall.RawConn, error) {
	return c.conn.SyscallConn()
}

type interConn struct {
	stream *quic.Stream
	local  net.Addr
	remote net.Addr
}

func (c *interConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

func (c *interConn) WriteMultiBuffer(mb buf.MultiBuffer) error {
	mb = buf.Compact(mb)
	mb, err := buf.WriteMultiBuffer(c, mb)
	buf.ReleaseMulti(mb)
	return err
}

func (c *interConn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

func (c *interConn) Close() error {
	return c.stream.Close()
}

func (c *interConn) LocalAddr() net.Addr {
	return c.local
}

func (c *interConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *interConn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *interConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *interConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}
