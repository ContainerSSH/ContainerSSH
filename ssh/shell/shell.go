package shell

import (
	"containerssh/backend"
	"containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
)

type requestMsg struct {
}

type responseMsg struct {
	exitStatus uint32
}

func onShellRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	shell, err := session.RequestShell()
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

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onShellRequest(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
