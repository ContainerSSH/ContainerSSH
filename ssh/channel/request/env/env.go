package env

import (
	"github.com/containerssh/containerssh/audit"
	audit2 "github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"

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

func (e ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, _ ssh.Channel, session backend.Session, auditChannel *audit.Channel) {
	e.logger.DebugF("Set env request: %s=%s", request.(*requestMsg).Name, request.(*requestMsg).Value)
	auditChannel.Message(audit2.MessageType_ChannelRequestSetEnv, audit2.PayloadChannelRequestSetEnv{
		Name:  request.(*requestMsg).Name,
		Value: request.(*requestMsg).Value,
	})
	err := session.SetEnv(request.(*requestMsg).Name, request.(*requestMsg).Value)
	if err != nil {
		e.logger.DebugF("Failed env request (%s)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
