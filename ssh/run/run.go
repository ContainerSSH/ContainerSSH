package run

import (
	"containerssh/backend"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
)

type responseMsg struct {
	exitStatus uint32
}

func run(program string, channel ssh.Channel, session backend.Session) error {
	shell, err := session.RequestProgram(program)
	if err != nil {
		return err
	}

	var once sync.Once
	closeSession := func() {
		exitCode := session.GetExitCode()
		session.Close()

		if exitCode < 0 {
			log.Printf("invalid exit code (%d)", exitCode)
		}

		//Send the exit status before closing the session. No reply is sent.
		_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(responseMsg{
			exitStatus: uint32(exitCode),
		}))
		//Close the channel as described by the RFC
		_ = channel.Close()
	}
	go func() {
		_, _ = io.Copy(channel, shell.Stdout)
		once.Do(closeSession)
	}()
	go func() {
		_, _ = io.Copy(shell.Stdin, channel)
		once.Do(closeSession)
	}()
	return nil
}
