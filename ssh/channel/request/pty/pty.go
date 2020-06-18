package pty

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	request2 "github.com/janoszen/containerssh/ssh/channel/request"
	"github.com/sirupsen/logrus"
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

func onPtyRequest(request *requestMsg, session backend.Session) error {
	logrus.Trace(fmt.Sprintf("PTY request"))
	err := session.SetPty()
	if err != nil {
		return err
	}
	return session.Resize(uint(request.Columns), uint(request.Rows))
}

var RequestTypeHandler = request2.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request2.Reply, channel ssh.Channel, session backend.Session) {
		err := onPtyRequest(request.(*requestMsg), session)
		if err != nil {
			logrus.Tracef("Failed pty request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
