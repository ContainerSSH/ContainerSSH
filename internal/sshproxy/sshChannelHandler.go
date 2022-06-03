package sshproxy

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/containerssh/libcontainerssh/internal/sshserver"
	ssh2 "github.com/containerssh/libcontainerssh/internal/ssh"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"golang.org/x/crypto/ssh"
)

type sshChannelHandler struct {
	lock              *sync.Mutex
	backingChannel     ssh.Channel
	connectionHandler *sshConnectionHandler
	requests           <-chan *ssh.Request
	session            sshserver.SessionChannel
	started            bool
	logger             log.Logger
	done               chan struct{}
	exited             bool
	ssh               *sshConnectionHandler
}

func (s *sshChannelHandler) handleBackendClientRequests(
	requests <-chan *ssh.Request,
	session sshserver.SessionChannel,
) {
	for {
		var request *ssh.Request
		var ok bool
		select {
		case request, ok = <-requests:
			if !ok {
				s.lock.Lock()
				s.logger.Debug(message.NewMessage(message.MSSHProxyBackendSessionClosed, "Backend closed session."))
				if err := s.session.Close(); err != nil && !errors.Is(err, io.EOF) {
					s.logger.Debug(message.Wrap(err,
						message.ESSHProxySessionCloseFailed, "Failed to close client-facing session after backend closed session."))
				}
				s.lock.Unlock()
				return
			}
		case <-s.done:
			return
		}
		s.lock.Lock()

		switch request.Type {
		case "exit-status":
			s.handleExitStatusFromBackend(request, session)
		case "exit-signal":
			s.handleExitSignalFromBackend(request, session)
		default:
			if request.WantReply {
				_ = request.Reply(false, []byte{})
			}
		}
		s.lock.Unlock()
	}
}

func (s *sshChannelHandler) handleExitStatusFromBackend(request *ssh.Request, session sshserver.SessionChannel) {
	exitStatus := &exitStatusPayload{}
	if err := ssh.Unmarshal(request.Payload, exitStatus); err != nil {
		s.logger.Debug(message.Wrap(err,
			message.MSSHProxyExitStatusDecodeFailed, "Received exit status from backend, but failed to decode payload."))
		if request.WantReply {
			_ = request.Reply(false, []byte("Failed to decode message."))
		}
	} else {
		s.logger.Debug(message.NewMessage(message.MSSHProxyExitStatus, "Received exit status from backend: %d", exitStatus.ExitStatus))
		session.ExitStatus(
			exitStatus.ExitStatus,
		)
		if request.WantReply {
			_ = request.Reply(true, []byte{})
		}
	}
}

func (s *sshChannelHandler) handleExitSignalFromBackend(request *ssh.Request, session sshserver.SessionChannel) {
	exitSignal := &exitSignalPayload{}
	if err := ssh.Unmarshal(request.Payload, exitSignal); err != nil {
		s.logger.Debug(message.Wrap(err,
			message.MSSHProxyExitSignalDecodeFailed, "Received exit signal from backend, but failed to decode payload."))
		if request.WantReply {
			_ = request.Reply(false, []byte{})
		}
	} else {
		s.logger.Debug(message.NewMessage(message.MSSHProxyExitSignal, "Received exit signal from backend: %s", exitSignal.Signal))
		session.ExitSignal(
			exitSignal.Signal,
			exitSignal.CoreDumped,
			exitSignal.ErrorMessage,
			exitSignal.LanguageTag,
		)
		if request.WantReply {
			_ = request.Reply(true, []byte{})
		}
	}
}

func (s *sshChannelHandler) streamStdio() error {
	if s.started {
		err := message.UserMessage(message.ESSHProxyProgramAlreadyStarted, "Cannot start new program after program has started.", "Client tried to start a program after the program was already started.")
		s.logger.Debug(err)
		return err
	}
	s.started = true
	go s.streamStdin()
	outWg := &sync.WaitGroup{}
	outWg.Add(2)
	go s.streamStdout(outWg)
	go s.streamStderr(outWg)
	go s.closeOnOutputComplete(outWg)
	return nil
}

func (s *sshChannelHandler) closeOnOutputComplete(outWg *sync.WaitGroup) {
	outWg.Wait()
	if err := s.session.CloseWrite(); err != nil && !errors.Is(err, io.EOF) {
		s.logger.Debug(
			message.NewMessage(
				message.ESSHProxyChannelCloseFailed,
				"Failed to close the SSH channel for writing.",
			),
		)
	}
}

func (s *sshChannelHandler) streamStderr(outWg *sync.WaitGroup) {
	if _, err := io.Copy(s.session.Stderr(), s.backingChannel.Stderr()); err != nil {
		if !errors.Is(err, io.EOF) {
			s.logger.Debug(message.Wrap(err, message.ESSHProxyStderrError, "Error copying stdout"))
		}
	}
	s.logger.Debug(message.NewMessage(message.MSSHProxyStderrComplete, "Stderr streaming complete."))
	outWg.Done()
}

func (s *sshChannelHandler) streamStdout(outWg *sync.WaitGroup) {
	if _, err := io.Copy(s.session.Stdout(), s.backingChannel); err != nil {
		if !errors.Is(err, io.EOF) {
			s.logger.Debug(message.Wrap(err, message.ESSHProxyStdoutError, "Error copying stdout"))
		}
	}
	s.logger.Debug(message.NewMessage(message.MSSHProxyStdoutComplete, "Stdout streaming complete."))
	outWg.Done()
}

func (s *sshChannelHandler) streamStdin() {
	if _, err := io.Copy(s.backingChannel, s.session.Stdin()); err != nil {
		if !errors.Is(err, io.EOF) {
			s.logger.Debug(message.Wrap(err, message.ESSHProxyStdinError, "Error copying stdin"))
		}
	}
	s.logger.Debug(message.NewMessage(message.MSSHProxySessionClose, "Stdin complete, closing backing session channel..."))
	if err := s.backingChannel.CloseWrite(); err != nil && !errors.Is(err, io.EOF) {
		s.logger.Debug(
			message.NewMessage(
				message.ESSHProxySessionCloseFailed,
				"Failed to close the backend SSH channel for writing.",
			),
		)
	}
	if err := s.backingChannel.Close(); err != nil && !errors.Is(err, io.EOF) {
		s.logger.Debug(
			message.NewMessage(
				message.ESSHProxySessionCloseFailed,
				"Failed to close the backend SSH channel.",
			),
		)
	} else {
		s.logger.Debug(message.NewMessage(message.MSSHProxySessionClosed, "Closed backing session channel."))
	}
}

func (s *sshChannelHandler) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {}

func (s *sshChannelHandler) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {
}

func (s *sshChannelHandler) sendRequest(name string, payload interface{}) error {
	var marshalledPayload []byte
	if payload != nil {
		marshalledPayload = ssh.Marshal(payload)
	}
	success, err := s.backingChannel.SendRequest(name, true, marshalledPayload)
	if err != nil {
		err := message.WrapUser(err,
			message.ESSHProxyBackendRequestFailed, "Cannot complete request.", "Sending a %s request to the backend resulted in an error.", name)
		s.logger.Debug(err)
		return err
	}
	if !success {
		err := message.UserMessage(message.ESSHProxyBackendRequestFailed, "Cannot complete request.", "Sending a %s request resulted in a rejection from the backend.", name)
		s.logger.Debug(err)
		return err
	}
	return nil
}

func (s *sshChannelHandler) OnEnvRequest(_ uint64, name string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		err := message.UserMessage(message.ESSHProxyProgramAlreadyStarted, "Cannot set environment variable after program has started.", "Client tried to set environment variables after the program was already started.")
		s.logger.Debug(err)
		return err
	}
	payload := envRequestPayload{
		Name:  name,
		Value: value,
	}
	return s.sendRequest("env", payload)
}

func (s *sshChannelHandler) OnPtyRequest(
	_ uint64,
	term string,
	columns uint32,
	rows uint32,
	width uint32,
	height uint32,
	modeList []byte,
) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		err := message.UserMessage(message.ESSHProxyProgramAlreadyStarted, "Cannot request PTY after program has started.", "Client tried request PTY after the program was already started.")
		s.logger.Debug(err)
		return err
	}
	payload := ptyRequestPayload{
		Term:     term,
		Columns:  columns,
		Rows:     rows,
		Width:    width,
		Height:   height,
		ModeList: modeList,
	}
	return s.sendRequest("pty-req", payload)
}

func (s *sshChannelHandler) OnExecRequest(_ uint64, program string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		err := message.UserMessage(
			message.ESSHProxyProgramAlreadyStarted,
			"Cannot start a program after another program has started.",
			"Client tried start a second program after the program was already started.",
		)
		s.logger.Debug(err)
		return err
	}
	payload := execRequestPayload{
		Exec: program,
	}
	err := s.sendRequest("exec", payload)
	if err != nil {
		return err
	}
	return s.streamStdio()
}

func (s *sshChannelHandler) OnShell(_ uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		err := message.UserMessage(
			message.ESSHProxyProgramAlreadyStarted,
			"Cannot start a program after another program has started.",
			"Client tried start a second program after the program was already started.",
		)
		s.logger.Debug(err)
		return err
	}
	err := s.sendRequest("shell", nil)
	if err != nil {
		return err
	}
	return s.streamStdio()
}

func (s *sshChannelHandler) OnSubsystem(_ uint64, subsystem string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.started {
		err := message.UserMessage(
			message.ESSHProxyProgramAlreadyStarted,
			"Cannot start a program after another program has started.",
			"Client tried start a second program after the program was already started.",
		)
		s.logger.Debug(err)
		return err
	}
	payload := subsystemRequestPayload{
		Subsystem: subsystem,
	}
	err := s.sendRequest("subsystem", payload)
	if err != nil {
		return err
	}
	return s.streamStdio()
}

func (s *sshChannelHandler) OnSignal(_ uint64, signal string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.started {
		err := message.UserMessage(
			message.ESSHProxyProgramNotStarted,
			"Cannot signal before program has started.",
			"Client tried send a signal before the program was started.",
		)
		s.logger.Debug(err)
		return err
	}
	payload := signalRequestPayload{
		Signal: signal,
	}
	return s.sendRequest("signal", payload)
}

func (s *sshChannelHandler) OnWindow(_ uint64, columns uint32, rows uint32, width uint32, height uint32) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.started {
		err := message.UserMessage(
			message.ESSHProxyProgramNotStarted,
			"Cannot resize window before program has started.",
			"Client tried request a window change before the program was started.",
		)
		s.logger.Debug(err)
		return err
	}
	payload := windowRequestPayload{
		Columns: columns,
		Rows:    rows,
		Width:   width,
		Height:  height,
	}
	if err := s.sendRequest("window-change", payload); err != nil {
		err := message.WrapUser(
			err,
			message.ESSHProxyWindowChangeFailed,
			"Cannot change window size.",
			"ContainerSSH cannot change the window size because of an error on the backend connection.",
		)
		s.logger.Debug(err)
		return err
	}
	return nil
}

func (s *sshChannelHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	payload := ssh2.X11RequestPayload{
		SingleConnection: singleConnection,
		Protocol: protocol,
		Cookie: cookie,
		Screen: screen,
	}
	if err := s.sendRequest("x11-req", payload); err != nil {
		err := message.WrapUser(
			err,
			message.ESSHProxyX11RequestFailed,
			"Error requesting X11 forwarding",
			"ContanerSSH cannot enable X11 forwarding because of an error on the backend connection",
		)
		s.logger.Debug(err)
		return err
	}
	return nil
}

func (s *sshChannelHandler) OnClose() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.started {
		s.logger.Debug(message.NewMessage(message.MSSHProxyBackendSessionClosing, "Client closed session before program start, closing backend session."))
		if err := s.backingChannel.Close(); err != nil {
			s.logger.Debug(message.Wrap(err, message.ESSHProxyBackendCloseFailed, "Failed to close backend channel."))
		}
	}

	s.logger.Debug(message.NewMessage(message.MSSHProxySessionClose, "Close received from client, closing backing channel."))
	if err := s.backingChannel.Close(); err != nil && !errors.Is(err, io.EOF) {
		s.logger.Debug(message.NewMessage(message.ESSHProxySessionCloseFailed, "Failed to close backing channel."))
	}
	close(s.done)
	s.exited = true
	s.ssh.networkHandler.wg.Done()
	s.logger.Debug(message.NewMessage(message.MSSHProxySessionClosed, "Backing channel closed."))
}

func (s *sshChannelHandler) OnShutdown(shutdownContext context.Context) {
	s.lock.Lock()
	s.logger.Debug(message.NewMessage(message.MSSHProxyShutdown, "Sending TERM signal on backing channel."))
	if err := s.sendRequest("signal", signalRequestPayload{
		Signal: "TERM",
	}); err != nil {
		s.logger.Debug(message.Wrap(err, message.ESSHProxySignalFailed, "Failed to deliver TERM signal to backend."))
	}
	s.lock.Unlock()

	select {
	case <-shutdownContext.Done():
		s.lock.Lock()
		if !s.exited {
			s.logger.Debug(message.NewMessage(message.MSSHProxyShutdown, "Sending KILL signal on backing channel."))
			if err := s.sendRequest("signal", signalRequestPayload{
				Signal: "KILL",
			}); err != nil {
				s.logger.Debug(message.Wrap(err,
					message.ESSHProxySignalFailed, "Failed to deliver KILL signal to backend."))
				_ = s.backingChannel.Close()
			}
		}
		s.lock.Unlock()
	case <-s.done:
	}
}
