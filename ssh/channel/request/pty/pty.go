package pty

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/format"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"

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

func (c ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, _ ssh.Channel, session backend.Session, auditChannel *audit.Channel) {
	c.logger.DebugF("PTY request")
	auditChannel.Message(format.MessageType_ChannelRequestPty, &format.PayloadChannelRequestPty{
		Columns: uint(request.(*requestMsg).Columns),
		Rows:    uint(request.(*requestMsg).Rows),
	})
	err := session.SetPty()
	if err != nil {
		c.logger.DebugF("failed PTY request (%v)", err)
		reply(false, nil)
		return
	}
	err = session.Resize(uint(request.(*requestMsg).Columns), uint(request.(*requestMsg).Rows))
	if err != nil {
		c.logger.DebugF("failed PTY request (%v)", err)
		reply(false, nil)
		return
	}

	reply(true, nil)
}
