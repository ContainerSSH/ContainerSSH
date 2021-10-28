package docker

import (
	"github.com/containerssh/libcontainerssh/internal/sshserver"
)

type sshConnectionHandler struct {
	sshserver.AbstractSSHConnectionHandler

	networkHandler *networkHandler
	username       string
	env            map[string]string
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {}

func (s *sshConnectionHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {}

func (s *sshConnectionHandler) OnSessionChannel(
	channelID uint64,
	_ []byte,
	session sshserver.SessionChannel,
) (
	channel sshserver.SessionChannelHandler,
	failureReason sshserver.ChannelRejection,
) {
	return &channelHandler{
		channelID:      channelID,
		networkHandler: s.networkHandler,
		username:       s.username,
		exitSent:       false,
		env:            map[string]string{},
		session:        session,
	}, nil
}
