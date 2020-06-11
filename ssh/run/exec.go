package run

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type execRequestMsg struct {
	Exec string
}

func onExecRequest(request *execRequestMsg, channel ssh.Channel, session backend.Session) error {
	logrus.Trace(fmt.Sprintf("Exec request: %s", request.Exec))
	return run(request.Exec, channel, session)
}

var ExecRequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &execRequestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onExecRequest(request.(*execRequestMsg), channel, session)
		if err != nil {
			logrus.Tracef("Failed exec request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
