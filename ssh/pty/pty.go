package pty

import (
	"containerssh/backend"
	"containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

func onPtyRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	err := session.SetPty()
	if err != nil {
		return err
	}
	return session.Resize(uint(request.Columns), uint(request.Rows))
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onPtyRequest(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
