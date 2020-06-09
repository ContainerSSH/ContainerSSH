package subsystem

import (
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/ssh/request"
	"golang.org/x/crypto/ssh"
	"log"
)

type requestMsg struct {
	Subsystem string
}

type responseMsg struct {
	exitStatus uint32
}

func onSubsystemRequest(request *requestMsg, channel ssh.Channel, session backend.Session) error {
	closeSession := func() {
		exitCode := session.GetExitCode()
		session.Close()

		if exitCode < 0 {
			log.Printf("invalid exit code (%d)", exitCode)
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
		log.Print(err)
		return err
	}
	return nil
}

var RequestTypeHandler = request.TypeHandler{
	GetRequestObject: func() interface{} { return &requestMsg{} },
	HandleRequest: func(request interface{}, reply request.Reply, channel ssh.Channel, session backend.Session) {
		err := onSubsystemRequest(request.(*requestMsg), channel, session)
		if err != nil {
			reply(false, nil)
		} else {
			reply(true, nil)
		}
	},
}
