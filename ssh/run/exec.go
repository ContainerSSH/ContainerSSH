package run

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
)

type execRequestMsg struct {
	Exec string
}

func onExecRequest(request *execRequestMsg, channel ssh.Channel, session backend.Session) error {
	return run(request.Exec, channel, session)
}

var ExecRequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &execRequestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onExecRequest(request.(*execRequestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
