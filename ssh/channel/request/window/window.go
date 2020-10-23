package window

import (
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"

	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

type ChannelRequestHandler struct {
	logger log.Logger
}

func New(logger log.Logger) channelRequest.TypeHandler {
	return &ChannelRequestHandler{
		logger: logger,
	}
}

func (e ChannelRequestHandler) GetRequestObject() interface{} {
	return &requestMsg{}
}

func (e ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, _ ssh.Channel, session backend.Session) {
	e.logger.DebugF("window change request: %dx%d", request.(*requestMsg).Rows, request.(*requestMsg).Columns)
	err := session.Resize(uint(request.(*requestMsg).Columns), uint(request.(*requestMsg).Rows))
	if err != nil {
		e.logger.DebugF("failed window change request (%v)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
