package shell

import (
	"github.com/containerssh/containerssh/audit"
	audit2 "github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"
	"github.com/containerssh/containerssh/ssh/channel/request/util"

	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
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

func (c ChannelRequestHandler) HandleRequest(_ interface{}, reply channelRequest.Reply, channel ssh.Channel, session backend.Session, auditChannel *audit.Channel) {
	c.logger.DebugF("shell request")
	auditChannel.Message(audit2.MessageType_ChannelRequestShell, audit2.PayloadChannelRequestShell{})
	err := util.Run("", channel, session, c.logger, auditChannel)
	if err != nil {
		c.logger.DebugF("failed exec request (%s)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
