package sshserver

import (
	"fmt"

	message2 "go.containerssh.io/libcontainerssh/message"
	"golang.org/x/crypto/ssh"
)

// region Factories

// NewChannelRejection constructs a Message that rejects a channel.
//
// Reason is the SSH rejection reason.
//
// Code is an error code allowing an administrator to identify the error that happened.
//
// UserMessage is the message that can be printed to the user if needed.
//
// Explanation is the explanation string to the system administrator. This is an fmt.Sprintf-compatible string
//
// Args are the arguments to Explanation to create a formatted message. It is recommended that these arguments also
// be added as labels to allow system administrators to index the error properly.
func NewChannelRejection(
	Reason ssh.RejectionReason,
	Code string,
	UserMessage string,
	Explanation string,
	Args ...interface{},
) ChannelRejection {
	return &message{
		reason:      Reason,
		code:        Code,
		userMessage: UserMessage,
		explanation: fmt.Sprintf(Explanation, Args...),
		labels:      map[message2.LabelName]message2.LabelValue{},
	}
}

// endregion

// region Message implementation

type message struct {
	code        string
	userMessage string
	explanation string
	labels      message2.Labels
	reason      ssh.RejectionReason
}

func (m *message) Reason() ssh.RejectionReason {
	return m.reason
}

func (m *message) Code() string {
	return m.code
}

func (m *message) UserMessage() string {
	return m.userMessage
}

func (m *message) Explanation() string {
	return m.explanation
}

func (m *message) Labels() message2.Labels {
	return m.labels
}

func (m *message) String() string {
	return m.userMessage
}

func (m *message) Label(name message2.LabelName, value message2.LabelValue) message2.Message {
	m.labels[name] = value
	return m
}

func (m *message) Error() string {
	return m.explanation
}

// endregion
