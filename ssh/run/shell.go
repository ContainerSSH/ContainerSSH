package run

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type ShellRequestMsg struct {
}

func onShellRequest(channel ssh.Channel, session backend.Session) error {
	logrus.Trace(fmt.Sprintf("Shell request"))
	return run("", channel, session)
}

var ShellRequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &ShellRequestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onShellRequest(channel, session)
		if err != nil {
			logrus.Tracef("Failed shell request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
