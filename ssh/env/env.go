package env

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Name  string
	Value string
}

func onSetEnvRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	return session.SetEnv(request.Name, request.Value)
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onSetEnvRequest(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
