package sshproxy

import (
	"context"
	"errors"
	"io"
	"sync"

	ssh2 "github.com/containerssh/libcontainerssh/internal/ssh"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
	"golang.org/x/crypto/ssh"
)

type sshConnectionHandler struct {
	networkHandler *networkConnectionHandler
	sshConn        ssh.Conn
	logger         log.Logger

	forwardMu      sync.Mutex
	reverseHandler sshserver.ReverseForward
}

func (s *sshConnectionHandler) handleRequests(requests <-chan *ssh.Request) {
	for {
		req, ok := <-requests
		if !ok {
			return
		}
		switch req.Type {
		case "keepalive@openssh.com":
			_ = req.Reply(false, nil)
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (s *sshConnectionHandler) handleChannels(newChannels <-chan ssh.NewChannel) {
	for {
		newChannel, ok := <-newChannels
		if !ok {
			return
		}
		switch newChannel.ChannelType() {
		case "x11":
			s.handleX11Channel(newChannel)
		case "forwarded-streamlocal@openssh.com":
			s.handleStreamLocalChannel(newChannel)
		case "forwarded-tcpip":
			s.handleReverseForwardChannel(newChannel)
		default:
			_ = newChannel.Reject(ssh.Prohibited, "Unsupported channel type")
		}
	}
}

func (s *sshConnectionHandler) rejectAllRequests(req <-chan *ssh.Request) {
	for {
		req, ok := <-req
		if !ok {
			return
		}
		if req.WantReply {
			_ = req.Reply(false, nil)
		}
	}
}

func (s *sshConnectionHandler) handleForward(closer *sync.Once, reader io.ReadCloser, writer io.WriteCloser) {
	_, _ = io.Copy(writer, reader)
	closer.Do(func() {
		s.logger.Debug("Closing forward channel")
		_ = writer.Close()
		_ = reader.Close()
	})
}

func (s *sshConnectionHandler) handleReverseForwardChannel(newChannel ssh.NewChannel) {
	var payload ssh2.ForwardTCPChannelOpenPayload
	err := ssh.Unmarshal(newChannel.ExtraData(), &payload)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyPayloadUnmarshalFailed,
			"Failed to open new reverse forwarding channel",
		)
		s.logger.Warning(m)
	}
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	clientChannel, _, err := s.reverseHandler.NewChannelTCP(payload.ConnectedAddress, payload.ConnectedPort, payload.OriginatorAddress, payload.OriginatorPort)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to open reverse forwarding channel to the client",
		)
		s.logger.Info(m)
	}
	serverChannel, req, err := newChannel.Accept()
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to accept forwarding channel from server",
		)
		s.logger.Info(m)
	}
	go s.rejectAllRequests(req)
	once := sync.Once{}

	go s.handleForward(&once, serverChannel, clientChannel)
	go s.handleForward(&once, clientChannel, serverChannel)
}

func (s *sshConnectionHandler) handleStreamLocalChannel(newChannel ssh.NewChannel) {
	var payload ssh2.StreamLocalForwardRequestPayload
	err := ssh.Unmarshal(newChannel.ExtraData(), &payload)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyPayloadUnmarshalFailed,
			"Failed to open new X11 forwarding channel",
		)
		s.logger.Warning(m)
	}
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	clientChannel, _, err := s.reverseHandler.NewChannelUnix(payload.SocketPath)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to open X11 channel to the client",
		)
		s.logger.Info(m)
	}
	serverChannel, req, err := newChannel.Accept()
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to accept X11 channel from server",
		)
		s.logger.Info(m)
	}
	go s.rejectAllRequests(req)
	once := sync.Once{}

	go s.handleForward(&once, serverChannel, clientChannel)
	go s.handleForward(&once, clientChannel, serverChannel)
}

func (s *sshConnectionHandler) handleX11Channel(newChannel ssh.NewChannel) {
	var payload ssh2.X11ChanOpenRequestPayload
	err := ssh.Unmarshal(newChannel.ExtraData(), &payload)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyPayloadUnmarshalFailed,
			"Failed to open new X11 forwarding channel",
		)
		s.logger.Warning(m)
	}
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	clientChannel, _, err := s.reverseHandler.NewChannelX11(payload.OriginatorAddress, payload.OriginatorPort)
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to open X11 channel to the client",
		)
		s.logger.Info(m)
	}
	serverChannel, req, err := newChannel.Accept()
	if err != nil {
		m := message.Wrap(
			err,
			message.ESSHProxyBackendForwardFailed,
			"Failed to accept X11 channel from server",
		)
		s.logger.Info(m)
	}
	go s.rejectAllRequests(req)
	once := sync.Once{}

	go s.handleForward(&once, serverChannel, clientChannel)
	go s.handleForward(&once, clientChannel, serverChannel)
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {
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
	backingChannel, requests, err := s.sshConn.OpenChannel("session", extraData)
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
		s.networkHandler.wg.Done()
		s.logger.Debug(failureReason)
		return nil, failureReason
	}

	sshChannelHandlerInstance := &sshChannelHandler{
		ssh:               s,
		lock:              &sync.Mutex{},
		backingChannel:    backingChannel,
		connectionHandler: s,
		requests:          requests,
		session:           session,
		logger:            s.logger,
		done:              make(chan struct{}),
	}
	go sshChannelHandlerInstance.handleBackendClientRequests(requests, session)

	s.logger.Debug(message.NewMessage(message.MSSHProxySessionOpen, "Session open on SSH backend..."))

	return sshChannelHandlerInstance, nil
}

func (s *sshConnectionHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	s.networkHandler.lock.Lock()
	if s.networkHandler.done {
		s.networkHandler.lock.Unlock()
		failureReason = sshserver.NewChannelRejection(
			ssh.ConnectionFailed,
			message.ESSHProxyShuttingDown,
			"Cannot open forward.",
			"Rejected new forwarding because connection is closing.",
		)
		s.logger.Debug(failureReason)
		return nil, failureReason
	}
	s.networkHandler.wg.Add(1)
	s.networkHandler.lock.Unlock()
	s.logger.Debug(message.NewMessage(message.MSSHProxyForward, "Opening new forwarding connection on SSH backend..."))
	payload := ssh2.ForwardTCPChannelOpenPayload{
		ConnectedAddress:  hostToConnect,
		ConnectedPort:     portToConnect,
		OriginatorAddress: originatorHost,
		OriginatorPort:    originatorPort,
	}
	mar := ssh.Marshal(payload)
	backingChannel, req, err := s.sshConn.OpenChannel("direct-tcpip", mar)
	if err != nil {
		realErr := &ssh.OpenChannelError{}
		if errors.As(err, &realErr) {
			failureReason = sshserver.NewChannelRejection(
				realErr.Reason,
				message.ESSHProxyBackendForwardFailed,
				realErr.Message,
				"Backend rejected channel with message: %s",
				realErr.Message,
			)
		} else {
			failureReason = sshserver.NewChannelRejection(
				ssh.ConnectionFailed,
				message.ESSHProxyBackendForwardFailed,
				"Cannot open session.",
				"Backend rejected channel with message: %s",
				err.Error(),
			)
		}
		s.logger.Debug(failureReason)
		return nil, failureReason
	}
	go s.rejectAllRequests(req)

	return backingChannel, nil
}

func (s *sshConnectionHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	payload := ssh2.ForwardTCPIPRequestPayload{
		Address: bindHost,
		Port:    bindPort,
	}
	mar := ssh.Marshal(payload)
	ok, _, err := s.sshConn.SendRequest(string(ssh2.RequestTypeReverseForward), true, mar)
	if err != nil {
		return err
	}
	if !ok {
		m := message.NewMessage(
			message.ESSHProxyBackendRequestFailed,
			"Failed to request forwarding because the backing SSH server rejected the request",
		)
		s.logger.Debug(m)
		return m
	}
	if s.reverseHandler == nil {
		s.reverseHandler = reverseHandler
	}
	return nil
}

func (s *sshConnectionHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	payload := ssh2.ForwardTCPIPRequestPayload{
		Address: bindHost,
		Port:    bindPort,
	}
	mar := ssh.Marshal(payload)
	ok, _, err := s.sshConn.SendRequest(string(ssh2.RequestTypeCancelReverseForward), true, mar)
	if err != nil {
		return err
	}
	if !ok {
		m := message.NewMessage(
			message.ESSHProxyBackendRequestFailed,
			"Failed to request forwarding cancellation because the backing SSH server rejected the request",
		)
		s.logger.Debug(m)
		return m
	}
	return nil
}

func (s *sshConnectionHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	payload := ssh2.DirectStreamLocalChannelOpenPayload{
		SocketPath: path,
	}
	mar := ssh.Marshal(payload)
	backingChannel, req, err := s.sshConn.OpenChannel(sshserver.ChannelTypeDirectStreamLocal, mar)
	if err != nil {
		realErr := &ssh.OpenChannelError{}
		if errors.As(err, &realErr) {
			failureReason = sshserver.NewChannelRejection(
				realErr.Reason,
				message.ESSHProxyBackendForwardFailed,
				realErr.Message,
				"Backend rejected channel with message: %s",
				realErr.Message,
			)
		} else {
			failureReason = sshserver.NewChannelRejection(
				ssh.ConnectionFailed,
				message.ESSHProxyBackendForwardFailed,
				"Cannot open session.",
				"Backend rejected channel with message: %s",
				err.Error(),
			)
		}
		s.logger.Debug(failureReason)
		return nil, failureReason
	}
	go s.rejectAllRequests(req)

	return backingChannel, nil
}

func (s *sshConnectionHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	s.forwardMu.Lock()
	defer s.forwardMu.Unlock()
	payload := ssh2.StreamLocalForwardRequestPayload{
		SocketPath: path,
	}
	mar := ssh.Marshal(payload)
	ok, _, err := s.sshConn.SendRequest(string(ssh2.RequestTypeStreamLocalForward), true, mar)
	if err != nil {
		return err
	}
	if !ok {
		m := message.NewMessage(
			message.ESSHProxyBackendRequestFailed,
			"Failed to request streamlocal because the backing SSH server rejected the request",
		)
		s.logger.Debug(m)
		return m
	}
	if s.reverseHandler == nil {
		s.reverseHandler = reverseHandler
	}
	return nil
}

func (s *sshConnectionHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	payload := ssh2.StreamLocalForwardRequestPayload{
		SocketPath: path,
	}
	mar := ssh.Marshal(payload)
	ok, _, err := s.sshConn.SendRequest(string(ssh2.RequestTypeCancelStreamLocalForward), true, mar)
	if err != nil {
		return err
	}
	if !ok {
		m := message.NewMessage(
			message.ESSHProxyBackendRequestFailed,
			"Failed to request streamlocal cancellation because the backing SSH server rejected the request",
		)
		s.logger.Debug(m)
		return m
	}
	return nil
}

func (s *sshConnectionHandler) OnShutdown(_ context.Context) {
}
