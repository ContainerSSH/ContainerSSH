package sshserver

import (
	"context"
)

type testSSHHandler struct {
	AbstractSSHConnectionHandler

	rootHandler    *testHandlerImpl
	networkHandler *testNetworkHandlerImpl
	username       string
	shutdown       bool
}

func (t *testSSHHandler) OnSessionChannel(_ uint64, _ []byte, session SessionChannel) (
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
