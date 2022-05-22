package security

import (
	"context"
	"sync"

	config2 "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
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
