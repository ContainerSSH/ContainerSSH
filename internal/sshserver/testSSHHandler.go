package sshserver

import (
	"context"

	"github.com/containerssh/libcontainerssh/metadata"
)

type testSSHHandler struct {
	AbstractSSHConnectionHandler

	rootHandler    *testHandlerImpl
	networkHandler *testNetworkHandlerImpl
	shutdown       bool
	metadata       metadata.ConnectionAuthenticatedMetadata
}

func (t *testSSHHandler) OnSessionChannel(_ metadata.ChannelMetadata, _ []byte, session SessionChannel) (
	channel SessionChannelHandler,
	failureReason ChannelRejection,
) {
	return &testSessionChannel{
		session: session,
		env:     map[string]string{},
	}, nil
}

func (t *testSSHHandler) OnShutdown(_ context.Context) {
	t.shutdown = true
}
