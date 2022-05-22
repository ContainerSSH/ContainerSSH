package kubernetes

import (
	"context"

	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/metadata"
)

type sshConnectionHandler struct {
	networkHandler *networkHandler
	username       string
	env            map[string]string
	files          map[string][]byte
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnShutdown(_ context.Context) {
}

func (s *sshConnectionHandler) OnSessionChannel(
	meta metadata.ChannelMetadata,
	_ []byte,
	session sshserver.SessionChannel,
) (
	channel sshserver.SessionChannelHandler,
	failureReason sshserver.ChannelRejection,
) {
	env := map[string]string{}
	for k, v := range s.env {
		env[k] = v
	}
	return &channelHandler{
		session:        session,
		channelID:      meta.ChannelID,
		networkHandler: s.networkHandler,
		username:       s.username,
		env:            env,
		files:          s.files,
	}, nil
}
