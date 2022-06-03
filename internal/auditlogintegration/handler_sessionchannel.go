package auditlogintegration

import (
	"context"

    "go.containerssh.io/libcontainerssh/internal/auditlog"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
)

type sessionChannelHandler struct {
	backend           sshserver.SessionChannelHandler
	connectionHandler *sshConnectionHandler
	audit             auditlog.Channel
	session           sshserver.SessionChannel
}

func (s *sessionChannelHandler) OnClose() {
	s.audit.OnClose()
	s.backend.OnClose()
}

func (s *sessionChannelHandler) OnShutdown(shutdownContext context.Context) {
	s.backend.OnShutdown(shutdownContext)
}

func (s *sessionChannelHandler) OnUnsupportedChannelRequest(requestID uint64, requestType string, payload []byte) {
	s.backend.OnUnsupportedChannelRequest(requestID, requestType, payload)
	s.audit.OnRequestUnknown(requestID, requestType, payload)
}

func (s *sessionChannelHandler) OnFailedDecodeChannelRequest(requestID uint64, requestType string, payload []byte, reason error) {
	s.backend.OnFailedDecodeChannelRequest(requestID, requestType, payload, reason)
	s.audit.OnRequestDecodeFailed(requestID, requestType, payload, reason.Error())
}

func (s *sessionChannelHandler) OnEnvRequest(requestID uint64, name string, value string) error {
	s.audit.OnRequestSetEnv(requestID, name, value)
	if err := s.backend.OnEnvRequest(requestID, name, value); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnExecRequest(
	requestID uint64,
	program string,
) error {
	s.audit.OnRequestExec(requestID, program)
	if err := s.backend.OnExecRequest(
		requestID,
		program,
	); err != nil {
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnPtyRequest(
	requestID uint64,
	term string,
	columns uint32,
	rows uint32,
	width uint32,
	height uint32,
	modeList []byte,
) error {
	s.audit.OnRequestPty(requestID, term, columns, rows, width, height, modeList)
	if err := s.backend.OnPtyRequest(requestID, term, columns, rows, width, height, modeList); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnShell(
	requestID uint64,
) error {
	s.audit.OnRequestShell(requestID)
	if err := s.backend.OnShell(
		requestID,
	); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnSignal(requestID uint64, signal string) error {
	s.audit.OnRequestSignal(requestID, signal)
	if err := s.backend.OnSignal(requestID, signal); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnSubsystem(
	requestID uint64,
	subsystem string,
) error {
	s.audit.OnRequestSubsystem(requestID, subsystem)
	if err := s.backend.OnSubsystem(
		requestID,
		subsystem,
	); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnWindow(
	requestID uint64,
	columns uint32,
	rows uint32,
	width uint32,
	height uint32,
) error {
	s.audit.OnRequestWindow(requestID, columns, rows, width, height)
	if err := s.backend.OnWindow(requestID, columns, rows, width, height); err != nil {
		s.audit.OnRequestFailed(requestID, err)
		return err
	}
	return nil
}

func (s *sessionChannelHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	s.audit.OnRequestX11(requestID, singleConnection, protocol, cookie, screen)

	return s.backend.OnX11Request(
		requestID,
		singleConnection,
		protocol,
		cookie,
		screen,
		&reverseHandlerProxy{
			backend: reverseHandler,
			connectionHandler: s.connectionHandler,
			channelType: sshserver.ChannelTypeX11,
		},
	)
}
