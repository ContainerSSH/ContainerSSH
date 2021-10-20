// +build plan9

package sshserver

import (
	"syscall"
)

func (s *serverImpl) socketControl(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
	})
}
