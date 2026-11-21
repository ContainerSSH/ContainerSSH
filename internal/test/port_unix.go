//go:build !windows && !plan9
// +build !windows,!plan9

package test

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func socketControl(network string, address string, conn syscall.RawConn) error {
	var reuseErr error
	if err := conn.Control(func(fd uintptr) {
		reuseErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	}); err != nil {
		return err
	}
	return reuseErr
}
