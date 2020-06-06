package shell

import (
	"containerssh/backend"
	"containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
)

type requestMsg struct {
}

func onShellRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	shell, err := session.RequestShell()
	if err != nil {
		return err
	}

	var once sync.Once
	closeBackendSession := func() {
		session.Close()
	}
	go func() {
		_, _ = io.Copy(channel, shell.Stdout)
		once.Do(closeBackendSession)
	}()
	go func() {
		_, _ = io.Copy(shell.Stdin, channel)
		once.Do(closeBackendSession)
	}()
	return nil
}

var ShellRequestTypeHandler = request.TypeHandler{
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
