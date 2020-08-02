package subsystem

import (
	"sync"

	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/log"
	channelRequest "github.com/janoszen/containerssh/ssh/channel/request"

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

func (c ChannelRequestHandler) HandleRequest(request interface{}, reply channelRequest.Reply, channel ssh.Channel, session backend.Session) {
	c.logger.DebugF("subsystem request: %s", request.(*requestMsg).Subsystem)
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
	err := session.RequestSubsystem(request.(*requestMsg).Subsystem, channel, channel, channel.Stderr(), closeSession)
	if err != nil {
		c.logger.DebugF("failed subsystem request (%v)", err)
		reply(false, nil)
	} else {
		reply(true, nil)
	}
}