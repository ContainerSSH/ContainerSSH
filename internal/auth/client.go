package auth

import (
	"github.com/containerssh/libcontainerssh/auth"
	"net"
)

// AuthenticationContext holds the results of an authentication.
type AuthenticationContext interface {
	// Success must return true or false of the authentication was successful / unsuccessful.
	Success() bool
	// Error returns the error that happened during the authentication.
	Error() error
	// Metadata returns a set of metadata entries that have been obtained during the authentication.
	Metadata() *auth.ConnectionMetadata
	// OnDisconnect is called when the client disconnects, or if the authentication fails due to a different reason.
	OnDisconnect()
}

// Client is an authentication client that provides authentication methods. Each authentication method returns a bool
// if the authentication was successful, and an error if the authentication failed due to a connection error.
//
// The authentication methods may also return a set of key-value metadata entries that can be consumed by the
// configuration server.
type Client interface {
	// Password authenticates with a password from the client. It returns a bool if the authentication as successful
	// or not. If an error happened while contacting the authentication server it will return an error.
	Password(
		username string,
		password []byte,
		connectionID string,
		remoteAddr net.IP,
	) AuthenticationContext

	// PubKey authenticates with a public key from the client. It returns a bool if the authentication as successful
	// or not. If an error happened while contacting the authentication server it will return an error.
	PubKey(
		username string,
		pubKey string,
		connectionID string,
		remoteAddr net.IP,
	) AuthenticationContext

	// KeyboardInteractive is a method to post a series of questions to the user and receive answers.
	KeyboardInteractive(
		username string,
		challenge func(
			instruction string,
			questions KeyboardInteractiveQuestions,
		) (answers KeyboardInteractiveAnswers, err error),
		connectionID string,
		remoteAddr net.IP,
	) AuthenticationContext

	// GSSAPIConfig is a method to generate and retrieve a GSSAPIServer interface for GSSAPI authentication
	GSSAPIConfig(
		connectionId string,
		addr net.IP,
	) GSSAPIServer
}

// KeyboardInteractiveQuestions is a list of questions for keyboard-interactive authentication
type KeyboardInteractiveQuestions []KeyboardInteractiveQuestion

// KeyboardInteractiveQuestion contains a question issued to a user as part of the keyboard-interactive exchange.
type KeyboardInteractiveQuestion struct {
	// ID is an optional opaque ID that can be used to identify a question in an answer. Can be left empty.
	ID string
	// Question is the question text sent to the user.
	Question string
	// EchoResponse should be set to true to show the typed response to the user.
	EchoResponse bool
}

// KeyboardInteractiveAnswers is a set of answer to a keyboard-interactive challenge.
type KeyboardInteractiveAnswers struct {
	// KeyboardInteractiveQuestion is the original question that was answered.
	Answers map[string]string
}

// GSSAPIServer is the interface for GSSAPI authentication
type GSSAPIServer interface {
	AuthenticationContext

	// AcceptSecContext is the GSSAPI function to verify the tokens
	AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error)

	// VerifyMIC is the GSSAPI function to verify the MIC (Message Integrity Code)
	VerifyMIC(micField []byte, micToken []byte) error

	// DeleteSecContext is the GSSAPI function to free all resources bound as part of an authentication attempt
	DeleteSecContext() error

	// AllowLogin is the authorization function. The username parameter
	// specifies the user that the authenticated user is trying to log in
	// as. Note! This is different from the gossh AllowLogin function in
	// which the username field is the authenticated username.
	AllowLogin(username string) error
}
