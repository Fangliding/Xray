package http

import (
	"io"
	"net"
	"time"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/signal/done"
)

type Connection struct {
	reader io.ReadCloser
	writer io.WriteCloser
	done   *done.Instance
	local  net.Addr
	remote net.Addr

	readDeadline  func(time.Time) error
	writeDeadline func(time.Time) error
}

func (c *Connection) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

// Write implements net.Conn.Write().
func (c *Connection) Write(b []byte) (int, error) {
	if c.done.Done() {
		return 0, io.ErrClosedPipe
	}

	return c.writer.Write(b)
}

// Close implements net.Conn.Close().
func (c *Connection) Close() error {
	common.Must(c.done.Close())
	common.Interrupt(c.reader)
	common.Close(c.writer)

	return nil
}

func (c *Connection) LocalAddr() net.Addr {
	return c.local
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.remote
}

func (c *Connection) SetDeadline(t time.Time) error {
	return nil
}

func (c *Connection) SetReadDeadline(t time.Time) error {
	if c.readDeadline != nil {
		return c.readDeadline(t)
	}
	return nil
}

func (c *Connection) SetWriteDeadline(t time.Time) error {
	if c.writeDeadline != nil {
		return c.writeDeadline(t)
	}
	return nil
}

type flushWriter struct {
	writer io.Writer
	closed *done.Instance
	flush  func() error
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	if fw.closed.Done() {
		return 0, io.ErrClosedPipe
	}

	n, err = fw.writer.Write(p)
	if err != nil {
		return n, err
	}
	fw.flush()
	return n, nil
}

func (fw flushWriter) Close() error {
	fw.closed.Close()
	common.Close(fw.writer)
	return nil
}

type waitReadCloser struct {
	io.ReadCloser
	ready *done.Instance
}

func (w *waitReadCloser) Set(rc io.ReadCloser) {
	w.ReadCloser = rc
	w.ready.Close()
}

func (w *waitReadCloser) Read(b []byte) (int, error) {
	if w.ReadCloser == nil {
		if <-w.ready.Wait(); w.ReadCloser == nil {
			return 0, io.ErrClosedPipe
		}
	}
	return w.ReadCloser.Read(b)
}

func (w *waitReadCloser) Close() error {
	if w.ReadCloser != nil {
		w.ReadCloser.Close()
	}
	w.ready.Close()
	return nil
}
