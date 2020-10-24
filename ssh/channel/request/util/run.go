package util

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/log"
	"sync"

	"github.com/containerssh/containerssh/backend"

	"golang.org/x/crypto/ssh"
)

type responseMsg struct {
	exitStatus uint32
}

func Run(program string, channel ssh.Channel, session backend.Session, logger log.Logger, auditChannel *audit.Channel) error {
	var mutex = &sync.Mutex{}
	closeSession := func() {
		mutex.Lock()
		session.Close()
		exitCode := session.GetExitCode()
		mutex.Unlock()

		if exitCode < 0 {
			logger.DebugF("invalid exit code (%d)", exitCode)
		}

		//Send the exit status before closing the session. No reply is sent.
		_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(responseMsg{
			exitStatus: uint32(exitCode),
		}))
		//Close the channel as described by the RFC
		_ = channel.Close()
	}
	stdIn, stdOut, stdErr := auditChannel.InterceptIo(channel, channel, channel.Stderr())
	err := session.RequestProgram(program, stdIn, stdOut, stdErr, closeSession)
	if err != nil {
		return err
	}
	return nil
}
