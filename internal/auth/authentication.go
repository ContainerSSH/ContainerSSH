// This file contains the interface definitions of the various authenticators supported in ContainerSSH.

package auth

import (
    "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/metadata"
)

// AuthenticationType is the root type for all authentication and authorization methods. This type is extended for each
// use case to limit the number of options.
type AuthenticationType string

// AuthenticationTypeAll indicates all authentication types are applicable.
const AuthenticationTypeAll AuthenticationType = ""

// AuthenticationTypePassword indicates the authentication where the user submits the password from their client.
const AuthenticationTypePassword AuthenticationType = "password"

// AuthenticationTypePublicKey indnicates the authentication type where the client performs a public-private key
// authentication.
const AuthenticationTypePublicKey AuthenticationType = "pubkey"

// AuthenticationTypeKeyboardInteractive indicates an authentication method where the user is asked a series of
// questions and must answer interactively.
const AuthenticationTypeKeyboardInteractive AuthenticationType = "keyboard-interactive"

// AuthenticationTypeGSSAPI indicates a cryptographic login method typically tied to the client computer, e.g.
// Kerberos.
const AuthenticationTypeGSSAPI AuthenticationType = "gssapi"

// AuthenticationTypeAuthz is the authorization after authentication method, which is intended to check
// if the user has permissions to log in. This does not make sense for all authentication methods.
const AuthenticationTypeAuthz AuthenticationType = "authz"

// PasswordAuthenticator validates the password of a user.
type PasswordAuthenticator interface {
	// Password authenticates with a password from the client. The returned AuthenticationContext contains the results
	// of the authentication process.
	//
	// - username is the username the user entered in their SSH session on the client side.
	// - password is the password the user entered when prompted for the password.
	// - connectionID is the identifier for the metadata
	Password(
		metadata metadata.ConnectionAuthPendingMetadata,
		password []byte,
	) AuthenticationContext
}

// PublicKeyAuthenticator authenticates using an SSH public key.
type PublicKeyAuthenticator interface {
	// PubKey authenticates with a public key from the client. The returned AuthenticationContext contains the results
	// of the authentication process.
	PubKey(
		metadata metadata.ConnectionAuthPendingMetadata,
		pubKey auth.PublicKey,
	) AuthenticationContext
}

// KeyboardInteractiveAuthenticator is authenticates using a question-and-answer back and forth between the client
// and the server.
type KeyboardInteractiveAuthenticator interface {
	// KeyboardInteractive is a method to post a series of questions to the user and receive answers.
	KeyboardInteractive(
		metadata metadata.ConnectionAuthPendingMetadata,
		challenge func(
			instruction string,
			questions KeyboardInteractiveQuestions,
		) (answers KeyboardInteractiveAnswers, err error),
	) AuthenticationContext
}

// GSSAPIAuthenticator authenticates using the GSSAPI method typically used for Kerberos authentication.
type GSSAPIAuthenticator interface {
	// GSSAPI is a method to generate and retrieve a GSSAPIServer interface for GSSAPI authentication.
	GSSAPI(
		metadata metadata.ConnectionMetadata,
	) GSSAPIServer
}

// AuthenticationContext holds the results of an authentication.
type AuthenticationContext interface {
	// Success must return true or false of the authentication was successful / unsuccessful.
	Success() bool
	// Error returns the error that happened during the authentication.
	Error() error
	// Metadata returns a set of metadata entries that have been obtained during the authentication.
	Metadata() metadata.ConnectionAuthenticatedMetadata
	// OnDisconnect is called when the client disconnects, or if the authentication fails due to a different reason.
	OnDisconnect()
}

// KeyboardInteractiveQuestions is a list of questions for keyboard-interactive authentication.
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

// GSSAPIServer is the interface for GSSAPI authentication. It extends ApplicationContext to provide the methods
// required for authenticating interactively.
type GSSAPIServer interface {
	// Success must return true or false of the authentication was successful / unsuccessful.
	Success() bool

	// Error returns the error that happened during the authentication.
	Error() error

	// AcceptSecContext is the GSSAPI function to verify the tokens.
	AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error)

	// VerifyMIC is the GSSAPI function to verify the MIC (Message Integrity Code).
	VerifyMIC(micField []byte, micToken []byte) error

	// DeleteSecContext is the GSSAPI function to free all resources bound as part of an authentication attempt.
	DeleteSecContext() error

	// AllowLogin is the authorization function. The username parameter
	// specifies the user that the authenticated user is trying to log in
	// as. Note! This is different from the gossh AllowLogin function in
	// which the username field is the authenticated username.
	AllowLogin(
		username string,
		meta metadata.ConnectionAuthPendingMetadata,
	) (metadata.ConnectionAuthenticatedMetadata, error)
}
