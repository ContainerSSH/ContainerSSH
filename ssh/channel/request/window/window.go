package window

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	request2 "github.com/janoszen/containerssh/ssh/channel/request"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

func onWindowChange(request *requestMsg, session backend.Session) error {
	log.Trace(fmt.Sprintf("Window change request: %dx%d", request.Rows, request.Columns))
	return session.Resize(uint(request.Columns), uint(request.Rows))
}

var RequestTypeHandler = request2.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request2.Reply, channel ssh.Channel, session backend.Session) {
		err := onWindowChange(request.(*requestMsg), session)
		if err != nil {
			log.Tracef("Failed window change request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
