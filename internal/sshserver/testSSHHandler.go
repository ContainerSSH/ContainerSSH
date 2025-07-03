package sshserver

import (
	"context"
	"fmt"

	"go.containerssh.io/containerssh/metadata"

	message2 "go.containerssh.io/containerssh/message"
	"golang.org/x/crypto/ssh"
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

func (s *testSSHHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel ForwardChannel, failureReason ChannelRejection) {
	return nil, NewChannelRejection(ssh.Prohibited, message2.ESSHNotImplemented, "Forwarding channel unimplemented in docker backend", "Forwarding channel unimplemented in docker backend")
}

func (s *testSSHHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testSSHHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testSSHHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel ForwardChannel, failureReason ChannelRejection) {
	return nil, NewChannelRejection(ssh.Prohibited, message2.ESSHNotImplemented, "Forwarding channel unimplemented in docker backend", "Forwarding channel unimplemented in docker backend")
}

func (s *testSSHHandler) OnRequestStreamLocal(
	path string,
	reverseHandler ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testSSHHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testSSHHandler) OnAuthAgentChannel(channelID uint64) (channel ForwardChannel, failureReason ChannelRejection) {
	return nil, NewChannelRejection(ssh.Prohibited, message2.ESSHNotImplemented, "SSH agent channel unimplemented in test backend", "SSH agent channel unimplemented in test backend")
}

func (t *testSSHHandler) OnRequestAuthAgent(
	agentChannel ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (t *testSSHHandler) OnShutdown(_ context.Context) {
	t.shutdown = true
}
