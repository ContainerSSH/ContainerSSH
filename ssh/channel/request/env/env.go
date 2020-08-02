package env

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/log"
	channelRequest "github.com/janoszen/containerssh/ssh/channel/request"

	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Name  string
	Value string
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

func (e ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, channel ssh.Channel, session backend.Session) {
	e.logger.DebugF("Set env request: %s=%s", request.(*requestMsg).Name, request.(*requestMsg).Value)
	err := session.SetEnv(request.(*requestMsg).Name, request.(*requestMsg).Value)
	if err != nil {
		e.logger.DebugF("Failed env request (%s)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
