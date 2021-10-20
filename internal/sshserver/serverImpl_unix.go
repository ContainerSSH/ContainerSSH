// +build !windows
// +build !plan9

package sshserver

import (
	"syscall"

	message2 "github.com/containerssh/containerssh/message"
)

func (s *serverImpl) socketControl(_, _ string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		err := syscall.SetsockoptInt(
			int(descriptor),
			syscall.SOL_SOCKET,
			syscall.SO_REUSEADDR,
			1,
		)
		if err != nil {
			s.logger.Warning(message2.NewMessage(ESOReuseFailed, "failed to set SO_REUSEADDR. Server may fail on restart"))
		}
	})
}
