//go:build windows
// +build windows

package sshserver

import (
	"syscall"

    messageCodes "go.containerssh.io/libcontainerssh/message"
)

func (s *serverImpl) socketControl(_, _ string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		err := syscall.SetsockoptInt(
			syscall.Handle(descriptor),
			syscall.SOL_SOCKET,
			syscall.SO_REUSEADDR,
			1,
		)
		if err != nil {
			s.logger.Warning(messageCodes.NewMessage(messageCodes.ESSHSOReuseFailed, "failed to set SO_REUSEADDR. Server may fail on restart"))
		}
	})
}
