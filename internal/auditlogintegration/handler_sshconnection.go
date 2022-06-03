package auditlogintegration

import (
	"context"
	"io"

    "go.containerssh.io/libcontainerssh/auditlog/message"
    "go.containerssh.io/libcontainerssh/internal/auditlog"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/metadata"
)

type sshConnectionHandler struct {
	backend sshserver.SSHConnectionHandler
	audit   auditlog.Connection
}

func (s *sshConnectionHandler) OnShutdown(shutdownContext context.Context) {
	s.backend.OnShutdown(shutdownContext)
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(requestID uint64, requestType string, payload []byte) {
	//todo audit payload
	s.audit.OnGlobalRequestUnknown(requestType)
	s.backend.OnUnsupportedGlobalRequest(requestID, requestType, payload)
}

func (s *sshConnectionHandler) OnFailedDecodeGlobalRequest(requestID uint64, requestType string, payload []byte, reason error) {
	s.audit.OnGlobalRequestDecodeFailed(requestID, requestType, payload, reason)
	s.backend.OnFailedDecodeGlobalRequest(requestID, requestType, payload, reason)
}

func (s *sshConnectionHandler) OnUnsupportedChannel(channelID uint64, channelType string, extraData []byte) {
	//todo audit extraData
	s.audit.OnNewChannelFailed(message.MakeChannelID(channelID), channelType, "unsupported channel type")
	s.backend.OnUnsupportedChannel(channelID, channelType, extraData)
}

func (s *sshConnectionHandler) OnSessionChannel(
	meta metadata.ChannelMetadata,
	extraData []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	proxy := &sessionProxy{
		backend: session,
	}
	backend, err := s.backend.OnSessionChannel(meta, extraData, proxy)
	if err != nil {
		return nil, err
	}
	auditChannel := s.audit.OnNewChannelSuccess(message.MakeChannelID(meta.ChannelID), "session")
	proxy.audit = auditChannel
	return &sessionChannelHandler{
		backend:           backend,
		connectionHandler: s,
		audit:             auditChannel,
		session:           session,
	}, nil
}

func (s *sshConnectionHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	s.audit.OnTCPForwardChannel(message.MakeChannelID(channelID), hostToConnect, portToConnect, originatorHost, originatorPort)

	backend, err := s.backend.OnTCPForwardChannel(channelID, hostToConnect, portToConnect, originatorHost, originatorPort)
	if err != nil {
		return nil, err
	}
	auditChannel := s.audit.OnNewChannelSuccess(message.MakeChannelID(channelID), "direct-tcpip")
	forwardProxy := auditChannel.GetForwardingProxy(backend)
	return forwardProxy, nil
}

func (s *sshConnectionHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	s.audit.OnRequestTCPReverseForward(bindHost, bindPort)
	return s.backend.OnRequestTCPReverseForward(
		bindHost,
		bindPort,
		&reverseHandlerProxy{
			backend:           reverseHandler,
			connectionHandler: s,
			channelType:       sshserver.ChannelTypeReverseForward,
		},
	)
}

func (s *sshConnectionHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	s.audit.OnRequestCancelTCPReverseForward(bindHost, bindPort)
	return s.backend.OnRequestCancelTCPReverseForward(bindHost, bindPort)
}

func (s *sshConnectionHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	s.audit.OnDirectStreamLocal(message.MakeChannelID(channelID), path)
	channel, err := s.backend.OnDirectStreamLocal(channelID, path)
	if err != nil {
		return nil, err
	}
	auditChannel := s.audit.OnNewChannelSuccess(message.MakeChannelID(channelID), sshserver.ChannelTypeDirectStreamLocal)
	return auditChannel.GetForwardingProxy(channel), nil
}

func (s *sshConnectionHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	s.audit.OnRequestStreamLocal(path)
	return s.backend.OnRequestStreamLocal(
		path,
		&reverseHandlerProxy{
			backend:           reverseHandler,
			connectionHandler: s,
			channelType:       sshserver.ChannelTypeForwardedStreamLocal,
		},
	)
}

func (s *sshConnectionHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	s.audit.OnRequestCancelStreamLocal(path)
	return s.backend.OnRequestCancelStreamLocal(path)
}

type reverseHandlerProxy struct {
	backend           sshserver.ReverseForward
	connectionHandler *sshConnectionHandler
	channelType       string
}

func (r *reverseHandlerProxy) NewChannelTCP(connectedAddress string, connectedPort uint32, originatorAddress string, originatorPort uint32) (sshserver.ForwardChannel, uint64, error) {
	channel, id, err := r.backend.NewChannelTCP(connectedAddress, connectedPort, originatorAddress, originatorPort)
	if err != nil {
		return nil, 0, err
	}

	r.connectionHandler.audit.OnReverseForwardChannel(message.MakeChannelID(id), connectedAddress, connectedPort, originatorAddress, originatorPort)
	auditChannel := r.connectionHandler.audit.OnNewChannelSuccess(message.MakeChannelID(id), r.channelType)
	forwardProxy := auditChannel.GetForwardingProxy(channel)
	return forwardProxy, id, nil
}

func (r *reverseHandlerProxy) NewChannelUnix(path string) (sshserver.ForwardChannel, uint64, error) {
	channel, id, err := r.backend.NewChannelUnix(path)
	if err != nil {
		return nil, 0, err
	}
	r.connectionHandler.audit.OnReverseStreamLocalChannel(message.MakeChannelID(id), path)
	auditChannel := r.connectionHandler.audit.OnNewChannelSuccess(message.MakeChannelID(id), r.channelType)
	forwardProxy := auditChannel.GetForwardingProxy(channel)
	return forwardProxy, id, nil
}

func (r *reverseHandlerProxy) NewChannelX11(originatorAddress string, originatorPort uint32) (sshserver.ForwardChannel, uint64, error) {
	channel, id, err := r.backend.NewChannelX11(originatorAddress, originatorPort)
	if err != nil {
		return nil, 0, err
	}

	r.connectionHandler.audit.OnReverseX11ForwardChannel(message.MakeChannelID(id), originatorAddress, originatorPort)
	auditChannel := r.connectionHandler.audit.OnNewChannelSuccess(message.MakeChannelID(id), r.channelType)
	forwardProxy := auditChannel.GetForwardingProxy(channel)
	return forwardProxy, id, nil
}

type sessionProxy struct {
	backend sshserver.SessionChannel
	audit   auditlog.Channel
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

func (s *sessionProxy) Stdin() io.Reader {
	if s.audit == nil {
		panic("BUG: stdin requested before channel is open")
	}
	if s.stdin == nil {
		s.stdin = s.audit.GetStdinProxy(s.backend.Stdin())
	}
	return s.stdin
}

func (s *sessionProxy) Stdout() io.Writer {
	if s.audit == nil {
		panic("BUG: stdout requested before channel is open")
	}
	if s.stdout == nil {
		s.stdout = s.audit.GetStdoutProxy(s.backend.Stdout())
	}
	return s.stdout
}

func (s *sessionProxy) Stderr() io.Writer {
	if s.audit == nil {
		panic("BUG: stderr requested before channel is open")
	}
	if s.stderr == nil {
		s.stderr = s.audit.GetStderrProxy(s.backend.Stderr())
	}
	return s.stderr
}

func (s *sessionProxy) ExitStatus(code uint32) {
	if s.audit == nil {
		panic("BUG: exit status requested before channel is open")
	}
	s.audit.OnExit(code)
	s.backend.ExitStatus(code)
}

func (s *sessionProxy) ExitSignal(signal string, coreDumped bool, errorMessage string, languageTag string) {
	if s.audit == nil {
		panic("BUG: exit signal requested before channel is open")
	}
	s.audit.OnExitSignal(signal, coreDumped, errorMessage, languageTag)
	s.backend.ExitSignal(signal, coreDumped, errorMessage, languageTag)
}

func (s *sessionProxy) CloseWrite() error {
	if s.audit == nil {
		panic("BUG: write close requested before channel is open")
	}
	s.audit.OnWriteClose()
	return s.backend.CloseWrite()
}

func (s *sessionProxy) Close() error {
	// Audit logging is done via the session channel hook.
	return s.backend.Close()
}
