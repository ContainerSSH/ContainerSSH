package sshproxy

import (
	"context"
	"errors"
	"sync"

	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
	"golang.org/x/crypto/ssh"
)

type sshConnectionHandler struct {
	networkHandler *networkConnectionHandler
	sshConn        ssh.Conn
	newChannels    <-chan ssh.NewChannel
	requests       <-chan *ssh.Request
	cli            *ssh.Client
	logger         log.Logger
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnSessionChannel(
	_ metadata.ChannelMetadata,
	extraData []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	s.networkHandler.lock.Lock()
	if s.networkHandler.done {
		failureReason = sshserver.NewChannelRejection(
			ssh.ConnectionFailed,
			message.ESSHProxyShuttingDown,
			"Cannot open session.",
			"Rejected new session because connection is closing.",
		)
		s.networkHandler.lock.Unlock()
		return
	}
	s.networkHandler.wg.Add(1)
	s.networkHandler.lock.Unlock()
	s.logger.Debug(message.NewMessage(message.MSSHProxySession, "Opening new session on SSH backend..."))
	backingChannel, requests, err := s.cli.OpenChannel("session", extraData)
	if err != nil {
		realErr := &ssh.OpenChannelError{}
		if errors.As(err, &realErr) {
			failureReason = sshserver.NewChannelRejection(
				realErr.Reason,
				message.ESSHProxyBackendSessionFailed,
				realErr.Message,
				"Backend rejected channel with message: %s",
				realErr.Message,
			)
		} else {
			failureReason = sshserver.NewChannelRejection(
				ssh.ConnectionFailed,
				message.ESSHProxyBackendSessionFailed,
				"Cannot open session.",
				"Backend rejected channel with message: %s",
				err.Error(),
			)
		}
		s.logger.Debug(failureReason)
		return nil, failureReason
	}

	sshChannelHandlerInstance := &sshChannelHandler{
		ssh:            s,
		lock:           &sync.Mutex{},
		backingChannel: backingChannel,
		requests:       requests,
		session:        session,
		logger:         s.logger,
		done:           make(chan struct{}),
	}
	go sshChannelHandlerInstance.handleBackendClientRequests(requests, session)

	s.logger.Debug(message.NewMessage(message.MSSHProxySessionOpen, "Session open on SSH backend..."))

	return sshChannelHandlerInstance, nil
}

func (s *sshConnectionHandler) OnShutdown(_ context.Context) {
}
