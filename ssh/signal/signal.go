package signal

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	signal string
}

func onSignalRequest(request *requestMsg, session backend.Session) error {
	logrus.Trace(fmt.Sprintf("Signal request: %s", request.signal))
	//todo should the list of signals allowed be filtered?
	return session.SendSignal("SIG" + request.signal)
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onSignalRequest(request.(*requestMsg), session)
		if err != nil {
			logrus.Tracef("Failed signal request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
