package security

import (
	"context"
	"sync"

    config2 "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
	"golang.org/x/crypto/ssh"
)

type sshConnectionHandler struct {
	config       config2.SecurityConfig
	backend      sshserver.SSHConnectionHandler
	sessionCount uint
	lock         *sync.Mutex
	logger       log.Logger
}

func (s *sshConnectionHandler) OnShutdown(shutdownContext context.Context) {
	s.backend.OnShutdown(shutdownContext)
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(requestID uint64, requestType string, payload []byte) {
	s.backend.OnUnsupportedGlobalRequest(requestID, requestType, payload)
}

func (s *sshConnectionHandler) OnFailedDecodeGlobalRequest(requestID uint64, requestType string, payload []byte, reason error) {
	s.backend.OnFailedDecodeGlobalRequest(requestID, requestType, payload, reason)
}

func (s *sshConnectionHandler) OnUnsupportedChannel(channelID uint64, channelType string, extraData []byte) {
	s.backend.OnUnsupportedChannel(channelID, channelType, extraData)
}

func (s *sshConnectionHandler) OnSessionChannel(
	meta metadata.ChannelMetadata,
	extraData []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.config.MaxSessions > -1 && s.sessionCount >= uint(s.config.MaxSessions) {
		err := &ErrTooManySessions{
			labels: message.Labels(map[message.LabelName]message.LabelValue{}),
		}
		s.logger.Debug(err)
		return nil, err
	}
	backend, err := s.backend.OnSessionChannel(meta, extraData, session)
	if err != nil {
		return nil, err
	}
	s.sessionCount++
	return &sessionHandler{
		config:        s.config,
		backend:       backend,
		sshConnection: s,
		logger:        s.logger,
	}, nil
}

func (s *sshConnectionHandler) getPolicy(primary config2.SecurityExecutionPolicy) config2.SecurityExecutionPolicy {
	if primary != config2.ExecutionPolicyUnconfigured {
		return primary
	}
	if s.config.DefaultMode != config2.ExecutionPolicyUnconfigured {
		return s.config.DefaultMode
	}
	return config2.ExecutionPolicyEnable
}

func (s *sshConnectionHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	mode := s.getPolicy(s.config.Forwarding.ForwardingMode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := sshserver.NewChannelRejection(
			ssh.Prohibited,
			message.ESecurityForwardingRejected,
			"Forwarding is rejected",
			"Forwarding is rejected because it is disabled",
		)
		s.logger.Debug(err)
		return nil, err
	case config2.ExecutionPolicyFilter:
		err := sshserver.NewChannelRejection(
			ssh.Prohibited,
			message.ESecurityForwardingRejected,
			"Forwarding is rejected",
			"Forwarding is rejected because it is set to filtered and filtering is currently not supported",
		)
		s.logger.Debug(err)
		return nil, err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnTCPForwardChannel(channelID, hostToConnect, portToConnect, originatorHost, originatorPort)
	}
}

func (s *sshConnectionHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	mode := s.getPolicy(s.config.Forwarding.ReverseForwardingMode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecurityReverseForwardingRejected,
			"Reverse forwarding is rejected",
			"Reverse forwarding is rejected because it is disabled in the config",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		err := message.UserMessage(
			message.ESecurityReverseForwardingRejected,
			"Reverse forwarding is rejected",
			"Reverse forwarding is rejected because it is set to filter and filtering reverse forwarding requests is not supported",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnRequestTCPReverseForward(bindHost, bindPort, reverseHandler)
	}
}

func (s *sshConnectionHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return s.backend.OnRequestCancelTCPReverseForward(
		bindHost,
		bindPort,
	)
}

func (s *sshConnectionHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	mode := s.getPolicy(s.config.Forwarding.SocketForwardingMode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := sshserver.NewChannelRejection(
			ssh.Prohibited,
			message.ESecurityForwardingRejected,
			"StreamLocal forwarding is rejected",
			"StreamLocal forwarding is rejected because it is disabled",
		)
		s.logger.Debug(err)
		return nil, err
	case config2.ExecutionPolicyFilter:
		err := sshserver.NewChannelRejection(
			ssh.Prohibited,
			message.ESecurityForwardingRejected,
			"StreamLocal forwarding is rejected",
			"StreamLocal forwarding is rejected because it is set to filtered and filtering is currently not supported",
		)
		s.logger.Debug(err)
		return nil, err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnDirectStreamLocal(channelID, path)
	}
}

func (s *sshConnectionHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	mode := s.getPolicy(s.config.Forwarding.SocketListenMode)
	switch mode {
	case config2.ExecutionPolicyDisable:
		err := message.UserMessage(
			message.ESecurityReverseForwardingRejected,
			"Reverse socket forwarding is rejected",
			"Reverse socket forwarding is rejected because it is disabled in the config",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyFilter:
		err := message.UserMessage(
			message.ESecurityReverseForwardingRejected,
			"Reverse socket forwarding is rejected",
			"Reverse socket forwarding is rejected because it is set to filter and filtering reverse forwarding requests is not supported",
		)
		s.logger.Debug(err)
		return err
	case config2.ExecutionPolicyEnable:
		fallthrough
	default:
		return s.backend.OnRequestStreamLocal(path, reverseHandler)
	}
}

func (s *sshConnectionHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return s.backend.OnRequestCancelStreamLocal(path)
}

// ErrTooManySessions indicates that too many sessions were opened in the same connection.
type ErrTooManySessions struct {
	labels message.Labels
}

// Label adds a label to the message.
func (e *ErrTooManySessions) Label(name message.LabelName, value message.LabelValue) message.Message {
	e.labels[name] = value
	return e
}

// Code returns the error code.
func (e *ErrTooManySessions) Code() string {
	return message.ESecurityMaxSessions
}

// Labels returns the list of labels for this message.
func (e *ErrTooManySessions) Labels() message.Labels {
	return e.labels
}

// Error contains the error for the logs.
func (e *ErrTooManySessions) Error() string {
	return "Too many sessions."
}

// Explanation is the message intended for the administrator.
func (e *ErrTooManySessions) Explanation() string {
	return "The user has opened too many sessions."
}

// UserMessage contains a message intended for the user.
func (e *ErrTooManySessions) UserMessage() string {
	return "Too many sessions."
}

// String returns the string representation of this message.
func (e *ErrTooManySessions) String() string {
	return e.UserMessage()
}

// Message contains a message intended for the user.
func (e *ErrTooManySessions) Message() string {
	return "Too many sessions."
}

// Reason contains the rejection code.
func (e *ErrTooManySessions) Reason() ssh.RejectionReason {
	return ssh.ResourceShortage
}
