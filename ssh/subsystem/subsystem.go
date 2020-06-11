package subsystem

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"sync"
)

type requestMsg struct {
	Subsystem string
}

type responseMsg struct {
	exitStatus uint32
}

func onSubsystemRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	logrus.Trace(fmt.Sprintf("Subsystem request: %s", request.Subsystem))
	var mutex = &sync.Mutex{}
	closeSession := func() {
		mutex.Lock()
		session.Close()
		exitCode := session.GetExitCode()
		mutex.Unlock()

		if exitCode < 0 {
			log.Warnf("invalid exit code (%d)", exitCode)
		}

		//Send the exit status before closing the session. No reply is sent.
		_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(responseMsg{
			exitStatus: uint32(exitCode),
		}))
		//Close the channel as described by the RFC
		_ = channel.Close()
	}
	err := session.RequestSubsystem(request.Subsystem, channel, channel, channel.Stderr(), closeSession)
	if err != nil {
		return err
	}
	return nil
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onSubsystemRequest(request.(*requestMsg), channel, session)
		if err != nil {
			log.Tracef("Failed subsystem request (%s)", err)
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
