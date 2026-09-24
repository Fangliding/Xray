//go:build !linux

package buf

import (
	"io"
	"syscall"
)

func lazyAllocReadBuffer(r io.Reader, _ syscall.RawConn) (*Buffer, error) {
	return ReadBuffer(r)
}
