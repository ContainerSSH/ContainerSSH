package window

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

func onWindowChange(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	return session.Resize(uint(request.Columns), uint(request.Rows))
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onWindowChange(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
