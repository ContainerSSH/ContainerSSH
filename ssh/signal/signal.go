package signal

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	signal string
}

func onSignalRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	//todo should the list of signals allowed be filtered?
	return session.SendSignal("SIG" + request.signal)
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onSignalRequest(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
