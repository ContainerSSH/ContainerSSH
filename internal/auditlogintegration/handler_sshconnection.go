package auditlogintegration

import (
	"context"
	"io"

	"github.com/containerssh/libcontainerssh/auditlog/message"
	"github.com/containerssh/libcontainerssh/internal/auditlog"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/metadata"
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
		backend: backend,
		audit:   auditChannel,
		session: session,
	}, nil
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
