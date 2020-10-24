package subsystem

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/protocol"
	"sync"

	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"

	"golang.org/x/crypto/ssh"
)

type requestMsg struct {
	Subsystem string
}

type responseMsg struct {
	exitStatus uint32
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
	c.logger.DebugF("subsystem request: %s", request.(*requestMsg).Subsystem)
	auditChannel.Message(protocol.MessageType_ChannelRequestSubsystem, protocol.PayloadChannelRequestSubsystem{
		Subsystem: request.(*requestMsg).Subsystem,
	})
	var mutex = &sync.Mutex{}
	closeSession := func() {
		mutex.Lock()
		session.Close()
		exitCode := session.GetExitCode()
		mutex.Unlock()

		if exitCode < 0 {
			c.logger.DebugF("invalid exit code (%d)", exitCode)
		}

		//Send the exit status before closing the session. No reply is sent.
		_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(responseMsg{
			exitStatus: uint32(exitCode),
		}))
		//Close the channel as described by the RFC
		_ = channel.Close()
	}

	stdIn, stdOut, stdErr := auditChannel.InterceptIo(channel, channel, channel.Stderr())
	err := session.RequestSubsystem(request.(*requestMsg).Subsystem, stdIn, stdOut, stdErr, closeSession)
	if err != nil {
		c.logger.DebugF("failed subsystem request (%v)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}
