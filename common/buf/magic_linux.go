//go:build linux

package buf

import (
	"io"
	"syscall"
)

func lazyAllocReadBuffer(_ io.Reader, c syscall.RawConn) (buf *Buffer, err error) {
	c.Read(func(fd uintptr) bool {
		buf = New()
		buf.Extend(Size)
		b := buf.Bytes()

		n, sErr := syscall.Read(int(fd), b)
		if sErr != nil {
			buf.Release()
			buf = nil
			if sErr == syscall.EAGAIN || sErr == syscall.EWOULDBLOCK {
				return false
			}
			err = sErr
			return true
		}
		buf.Resize(0, int32(n))
		return true
	})
	if buf.Len() == 0 && err == nil {
		buf.Release()
		buf = nil
		err = io.EOF
	}
	return buf, err
}
