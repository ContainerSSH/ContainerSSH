package security //nolint:testpackage

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/metadata"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/ssh"
)

func TestMaxSessions(t *testing.T) {
	backend := &dummySSHBackend{
		exitChannel: make(chan struct{}),
	}
	ssh := &sshConnectionHandler{
		config: config.SecurityConfig{
			MaxSessions: 10,
		},
		backend: backend,
		lock:    &sync.Mutex{},
		logger:  log.NewTestLogger(t),
	}

	for i := 0; i < ssh.config.MaxSessions; i++ {
		handler, err := ssh.OnSessionChannel(createChannelMetadata(i), []byte{}, &sessionChannel{})
		assert.NoError(t, err)
		assert.NoError(
			t, handler.OnShell(0),
		)
	}
	_, err := ssh.OnSessionChannel(createChannelMetadata(ssh.config.MaxSessions), []byte{}, &sessionChannel{})
	assert.Error(t, err)
	for i := 0; i < ssh.config.MaxSessions; i++ {
		backend.exitChannel <- struct{}{}
	}
}

func createChannelMetadata(i int) metadata.ChannelMetadata {
	return metadata.ChannelMetadata{
		Connection: metadata.ConnectionAuthenticatedMetadata{
			ConnectionAuthPendingMetadata: metadata.ConnectionAuthPendingMetadata{
				ConnectionMetadata: metadata.ConnectionMetadata{
					RemoteAddress: metadata.RemoteAddress(
						net.TCPAddr{
							IP:   net.ParseIP("127.0.0.1"),
							Port: 22,
						},
					),
					ConnectionID: "asdf",
					Metadata:     nil,
					Environment:  nil,
					Files:        nil,
				},
				ClientVersion: "",
				Username:      "test",
			},
			AuthenticatedUsername: "test",
		},
		ChannelID: uint64(i),
	}
}

type sessionChannel struct {
}

func (s *sessionChannel) Stdin() io.Reader {
	panic("implement me")
}

func (s *sessionChannel) Stdout() io.Writer {
	panic("implement me")
}

func (s *sessionChannel) Stderr() io.Writer {
	panic("implement me")
}

func (s *sessionChannel) ExitStatus(_ uint32) {
	panic("implement me")
}

func (s *sessionChannel) ExitSignal(_ string, _ bool, _ string, _ string) {
	panic("implement me")
}

func (s *sessionChannel) CloseWrite() error {
	panic("implement me")
}

func (s *sessionChannel) Close() error {
	panic("implement me")
}

type dummySSHBackend struct {
	exitChannel chan struct{}
}

func (d *dummySSHBackend) OnShutdown(_ context.Context) {
	panic("implement me")
}

func (d *dummySSHBackend) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
	panic("implement me")
}

func (b *dummySSHBackend) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {
	panic("implement me")
}

func (d *dummySSHBackend) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {
	panic("implement me")
}

func (d *dummySSHBackend) OnSessionChannel(
	_ metadata.ChannelMetadata,
	_ []byte,
	_ sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	return &dummyBackend{
		exit: d.exitChannel,
	}, nil
}

func (s *dummySSHBackend) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented in docker backend", "Forwading channel unimplemented in docker backend")
}

func (s *dummySSHBackend) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummySSHBackend) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummySSHBackend) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented in docker backend", "Forwading channel unimplemented in docker backend")
}

func (s *dummySSHBackend) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummySSHBackend) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}
