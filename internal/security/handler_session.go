package security

import (
	"context"

    config2 "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
)

type sessionHandler struct {
	config        config2.SecurityConfig
	backend       sshserver.SessionChannelHandler
	sshConnection *sshConnectionHandler
	logger        log.Logger
}

func (s *sessionHandler) OnClose() {
	s.backend.OnClose()
}

func (s *sessionHandler) OnShutdown(shutdownContext context.Context) {
	s.backend.OnShutdown(shutdownContext)
}

func (s *sessionHandler) OnUnsupportedChannelRequest(requestID uint64, requestType string, payload []byte) {
	s.backend.OnUnsupportedChannelRequest(requestID, requestType, payload)
}

func (s *sessionHandler) OnFailedDecodeChannelRequest(
	requestID uint64,
	requestType string,
	payload []byte,
	reason error,
) {
	s.backend.OnFailedDecodeChannelRequest(requestID, requestType, payload, reason)
}

func (s *sessionHandler) getPolicy(primary config2.SecurityExecutionPolicy) config2.SecurityExecutionPolicy {
	if primary != config2.ExecutionPolicyUnconfigured {
		return primary
	}
	if s.config.DefaultMode != config2.ExecutionPolicyUnconfigured {
		return s.config.DefaultMode
	}
	return config2.ExecutionPolicyEnable
}

func (s *sessionHandler) contains(items []string, item string) bool {
	for _, searchItem := range items {
		if searchItem == item {
			return true
		}
	}
	return false
}

func (s *sessionHandler) OnEnvRequest(requestID uint64, name string, value string) error {
	mode := s.getPolicy(s.config.Env.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecurityEnvRejected,
			"Environment variable setting rejected.",
			"Setting an environment variable is rejected because it is disabled in the security settings.",
		).Label("name", name)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		if s.contains(s.config.Env.Allow, name) {
			return s.backend.OnEnvRequest(requestID, name, value)
		}
		err := message.UserMessage(
			message.ESecurityEnvRejected,
			"Environment variable setting rejected.",
			"Setting an environment variable is rejected because it does not match the allow list.",
		).Label("name", name)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		if !s.contains(s.config.Env.Deny, name) {
			return s.backend.OnEnvRequest(requestID, name, value)
		}
		err := message.UserMessage(
			message.ESecurityEnvRejected,
			"Environment variable setting rejected.",
			"Setting an environment variable is rejected because it matches the deny list.",
		).Label("name", name)
		s.logger.Debug(err)
		return err
	}
}

func (s *sessionHandler) OnPtyRequest(
	requestID uint64,
	term string,
	columns uint32,
	rows uint32,
	width uint32,
	height uint32,
	modeList []byte,
) error {
	mode := s.getPolicy(s.config.TTY.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		fallthrough
	case config2.ExecutionPolicyFilter:
		err := message.UserMessage(
			message.ESecurityTTYRejected,
			"TTY allocation disabled.",
			"TTY allocation is disabled in the security settings.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnPtyRequest(requestID, term, columns, rows, width, height, modeList)
	}
}

func (s *sessionHandler) OnExecRequest(
	requestID uint64,
	program string,
) error {
	mode := s.getPolicy(s.config.Command.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecurityExecRejected,
			"Command execution disabled.",
			"Command execution is disabled in the security settings.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		if !s.contains(s.config.Command.Allow, program) {
			err := message.UserMessage(
				message.ESecurityExecRejected,
				"Command execution disabled.",
				"The specified command passed from the client does not match the specified allow list.",
			)
			s.logger.Debug(err)
			return err
		}
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
	}
	if s.config.ForceCommand == "" {
		return s.backend.OnExecRequest(requestID, program)
	}
	if err := s.backend.OnEnvRequest(requestID, "SSH_ORIGINAL_COMMAND", program); err != nil {
		err := message.WrapUser(
			err,
			message.ESecurityFailedSetEnv,
			"Could not execute program.",
			"Command execution failed because the security layer could not set the SSH_ORIGINAL_COMMAND variable.",
		)
		s.logger.Error(err)
		return err
	}
	s.logger.Debug(
		message.NewMessage(
			message.MSecurityForcingCommand,
			"Forcing command execution to %s",
			s.config.ForceCommand,
		))
	return s.backend.OnExecRequest(requestID, s.config.ForceCommand)
}

func (s *sessionHandler) OnShell(
	requestID uint64,
) error {
	mode := s.getPolicy(s.config.Shell.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		fallthrough
	case config2.ExecutionPolicyFilter:
		err := message.UserMessage(
			message.ESecurityShellRejected,
			"Shell execution disabled.",
			"Shell execution is disabled in the security settings.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
	}
	if s.config.ForceCommand == "" {
		return s.backend.OnShell(requestID)
	}
	s.logger.Debug(
		message.NewMessage(
			message.MSecurityForcingCommand,
			"Forcing command execution to %s",
			s.config.ForceCommand,
		))
	return s.backend.OnExecRequest(requestID, s.config.ForceCommand)
}

func (s *sessionHandler) OnSubsystem(
	requestID uint64,
	subsystem string,
) error {
	mode := s.getPolicy(s.config.Subsystem.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecuritySubsystemRejected,
			"Subsystem execution disabled.",
			"Subsystem execution is disabled in the security settings.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		if !s.contains(s.config.Subsystem.Allow, subsystem) {
			err := message.UserMessage(
				message.ESecuritySubsystemRejected,
				"Subsystem execution disabled.",
				"The specified subsystem does not match the allowed subsystems list.",
			)
			s.logger.Debug(err)
			return err
		}
	case config2.ExecutionPolicyEnable:
		if s.contains(s.config.Subsystem.Deny, subsystem) {
			err := message.UserMessage(
				message.ESecuritySubsystemRejected,
				"Subsystem execution disabled.",
				"The subsystem execution is rejected because the specified subsystem matches the deny list.",
			)
			s.logger.Debug(err)
			return err
		}
	default:
	}
	if s.config.ForceCommand == "" {
		return s.backend.OnSubsystem(requestID, subsystem)
	}
	if err := s.backend.OnEnvRequest(requestID, "SSH_ORIGINAL_COMMAND", subsystem); err != nil {
		err := message.WrapUser(
			err,
			message.ESecurityFailedSetEnv,
			"Could not execute program.",
			"Command execution failed because the security layer could not set the SSH_ORIGINAL_COMMAND variable.",
		)
		s.logger.Error(err)
		return err
	}
	s.logger.Debug(
		message.NewMessage(
			message.MSecurityForcingCommand,
			"Forcing command execution to %s",
			s.config.ForceCommand,
		))
	return s.backend.OnExecRequest(requestID, s.config.ForceCommand)
}

func (s *sessionHandler) OnSignal(requestID uint64, signal string) error {
	mode := s.getPolicy(s.config.Shell.Mode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecuritySignalRejected,
			"Sending signals is rejected.",
			"Sending the signal is rejected because signal delivery is disabled.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		if s.contains(s.config.Signal.Allow, signal) {
			return s.backend.OnSignal(requestID, signal)
		}
		err := message.UserMessage(
			message.ESecuritySignalRejected,
			"Sending signals is rejected.",
			"Sending the signal is rejected because the specified signal does not match the allow list.",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		if !s.contains(s.config.Signal.Deny, signal) {
			return s.backend.OnSignal(requestID, signal)
		}
		err := message.UserMessage(
			message.ESecuritySignalRejected,
			"Sending signals is rejected.",
			"Sending the signal is rejected because the specified signal matches the deny list.",
		)
		s.logger.Debug(err)
		return err
	}
}

func (s *sessionHandler) OnWindow(requestID uint64, columns uint32, rows uint32, width uint32, height uint32) error {
	return s.backend.OnWindow(requestID, columns, rows, width, height)
}

func (s *sessionHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	mode := s.getPolicy(s.config.Forwarding.X11ForwardingMode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecurityX11ForwardingRejected,
			"X11 forwarding is rejected",
			"X11 forwarding is rejected because it is disabled in the config",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		err := message.UserMessage(
			message.ESecurityX11ForwardingRejected,
			"X11 forwarding is rejected",
			"X11 forwarding is rejected because it is set to filter and filterint X11 requests is not supported",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnX11Request(requestID, singleConnection, protocol, cookie, screen, reverseHandler)
	}
}
