package docker

import (
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/metadata"
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
	meta metadata.ChannelMetadata,
	_ []byte,
	session sshserver.SessionChannel,
) (
	channel sshserver.SessionChannelHandler,
	failureReason sshserver.ChannelRejection,
) {
	return &channelHandler{
		channelID:      meta.ChannelID,
		networkHandler: s.networkHandler,
		username:       s.username,
		exitSent:       false,
		env:            map[string]string{},
		session:        session,
	}, nil
}
