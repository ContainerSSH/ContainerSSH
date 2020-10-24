package exec

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/protocol"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"
	"github.com/containerssh/containerssh/ssh/channel/request/util"

	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Exec string
}

type ChannelRequestHandler struct {
	logger log.Logger
}

func New(logger log.Logger) channelRequest.TypeHandler {
	return &ChannelRequestHandler{
		logger: logger,
	}
}

func (c ChannelRequestHandler) GetRequestObject() interface{} {
	return &requestMsg{}
}

func (c ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, channel ssh.Channel, session backend.Session, auditChannel *audit.Channel) {
	c.logger.DebugF("Exec request: %s", request.(*requestMsg).Exec)
	auditChannel.Message(protocol.MessageType_ChannelRequestExec, protocol.MessageChannelRequestExec{
		Program: request.(*requestMsg).Exec,
	})
	err := util.Run(request.(*requestMsg).Exec, channel, session, c.logger, auditChannel)
	if err != nil {
		c.logger.DebugF("failed exec request (%s)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
