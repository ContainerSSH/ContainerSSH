package message

import (
	"fmt"
)

// region Factories

// UserMessage constructs a Message.
//
//   - Code is an error code allowing an administrator to identify the error that happened.
//   - UserMessage is the message that can be printed to the user if needed.
//   - Explanation is the explanation string to the system administrator. This is an fmt.Sprintf-compatible string
//   - Args are the arguments to Explanation to create a formatted message. It is recommended that these arguments also
//     be added as labels to allow system administrators to index the error properly.
func UserMessage(Code string, UserMessage string, Explanation string, Args ...interface{}) Message {
	return &message{
		code:        Code,
		userMessage: UserMessage,
		explanation: fmt.Sprintf(Explanation, Args...),
		labels:      map[LabelName]LabelValue{},
	}
}

// WrapUser creates a Message wrapping an error with a user-facing message.
//
//   - Cause is the original error that can be accessed with the Unwrap method.
//   - Code is an error code allowing an administrator to identify the error that happened.
//   - UserMessage is the message that can be printed to the user if needed.
//   - Explanation is the explanation string to the system administrator. This is an fmt.Sprintf-compatible string
//   - Args are the arguments to Explanation to create a formatted message. It is recommended that these arguments also
//     be added as labels to allow system administrators to index the error properly.
//
//goland:noinspection GoUnusedExportedFunction
func WrapUser(Cause error, Code string, User string, Explanation string, Args ...interface{}) WrappingMessage {
	return &wrappingMessage{
		Message: UserMessage(Code, User, Explanation+" (%v)", append(Args, Cause)...),
		cause:   Cause,
	}
}

// NewMessage creates an internal error with only the explanation for the administrator inserted.
//
//   - Code is an error code allowing an administrator to identify the error that happened.
//   - Explanation is the explanation string to the system administrator. This is an fmt.Sprintf-compatible string
//   - Args are the arguments to Explanation to create a formatted message. It is recommended that these arguments also
//     be added as labels to allow system administrators to index the error properly.
func NewMessage(Code string, Explanation string, Args ...interface{}) Message {
	return UserMessage(
		Code,
		"Internal Error",
		Explanation,
		Args...,
	)
}

// Wrap creates a wrapped error with a specific Code and Explanation string. The wrapping method will automatically
//
//	append the error message in brackets.
//
//   - Cause is the original error that can be accessed with the Unwrap method.
//   - Code is an error code allowing an administrator to identify the error that happened.
//   - Explanation is the explanation string to the system administrator. This is an fmt.Sprintf-compatible string
//   - Args are the arguments to Explanation to create a formatted message. It is recommended that these arguments also
//     be added as labels to allow system administrators to index the error properly.
func Wrap(Cause error, Code string, Explanation string, Args ...interface{}) WrappingMessage {
	return &wrappingMessage{
		Message: NewMessage(Code, Explanation+" (%v)", append(Args, Cause)...),
		cause:   Cause,
	}
}

// endregion

// region Interfaces

// Message is a message structure for error reporting in ContainerSSH. The actual implementations may differ, but we
// provide the UserMessage method to construct a message that conforms to this interface.
type Message interface {
	// Error is the Go-compatible error message.
	Error() string
	// String returns the string representation of this message.
	String() string

	// Code is a unique code identifying log messages.
	Code() string
	// UserMessage is a message that is safe to print/send to the end user.
	UserMessage() string
	// Explanation is the text explanation for the system administrator.
	Explanation() string
	// Labels are a set of extra labels for the message containing information.
	Labels() Labels
	// Label adds a label to the message.
	Label(name LabelName, value LabelValue) Message
}

// WrappingMessage is a message that wraps a different error.
type WrappingMessage interface {
	Message
	// Unwrap returns the original error wrapped in this message.
	Unwrap() error
}

// endregion

// region Message implementation

type message struct {
	code        string
	userMessage string
	explanation string
	labels      Labels
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

func (m *message) Labels() Labels {
	return m.labels
}

func (m *message) String() string {
	return m.userMessage
}

func (m *message) Label(name LabelName, value LabelValue) Message {
	m.labels[name] = value
	return m
}

func (m *message) Error() string {
	return m.explanation
}

// endregion

// region Wrapping message implementation

type wrappingMessage struct {
	Message
	cause error
}

func (w wrappingMessage) Unwrap() error {
	return w.cause
}

// endregion
