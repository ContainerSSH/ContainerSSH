package sshserver

import (
	"context"
	"fmt"
	"io"

	"github.com/containerssh/libcontainerssh/auth"
	auth2 "github.com/containerssh/libcontainerssh/internal/auth"
	message2 "github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"

	"golang.org/x/crypto/ssh"
)

// Handler is the basic conformanceTestHandler for SSH connections. It contains several methods to handle startup and operations of the
//         server
type Handler interface {
	// OnReady is called when the server is ready to receive connections. It has an opportunity to return an error to
	//         abort the startup.
	OnReady() error

	// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
	//            for the shutdown, after which the server should abort all running connections and return as fast as
	//            possible.
	OnShutdown(shutdownContext context.Context)

	// OnNetworkConnection is called when a new network connection is opened. It must either return a
	// NetworkConnectionHandler object or an error. In case of an error the network connection is closed.
	OnNetworkConnection(metadata.ConnectionMetadata) (NetworkConnectionHandler, metadata.ConnectionMetadata, error)
}

// AuthResponse is the result of the authentication process.
type AuthResponse uint8

const (
	// AuthResponseSuccess indicates that the authentication was successful.
	AuthResponseSuccess AuthResponse = 1

	// AuthResponseFailure indicates that the authentication failed for invalid credentials.
	AuthResponseFailure AuthResponse = 2

	// AuthResponseUnavailable indicates that the authentication could not be performed because a backend system failed
	//                         to respond.
	AuthResponseUnavailable AuthResponse = 3
)

// KeyboardInteractiveQuestion contains a question issued to a user as part of the keyboard-interactive exchange.
type KeyboardInteractiveQuestion struct {
	// ID is an optional opaque ID that can be used to identify a question in an answer. Can be left empty.
	ID string
	// Question is the question text sent to the user.
	Question string
	// EchoResponse should be set to true to show the typed response to the user.
	EchoResponse bool
}

func (k *KeyboardInteractiveQuestion) getID() string {
	if k.ID != "" {
		return k.ID
	}
	return k.Question
}

// KeyboardInteractiveQuestions is a list of questions for keyboard-interactive authentication
type KeyboardInteractiveQuestions []KeyboardInteractiveQuestion

func (k *KeyboardInteractiveQuestions) Add(question KeyboardInteractiveQuestion) {
	*k = append(*k, question)
}

// KeyboardInteractiveAnswers is a set of answer to a keyboard-interactive challenge.
type KeyboardInteractiveAnswers struct {
	// KeyboardInteractiveQuestion is the original question that was answered.
	answers map[string]string
}

// Get returns the answer for a question, or an error if no answer is present.
func (k *KeyboardInteractiveAnswers) Get(question KeyboardInteractiveQuestion) (string, error) {
	if val, ok := k.answers[question.getID()]; ok {
		return val, nil
	}
	return "", fmt.Errorf("no answer for question")
}

// GetByQuestionText returns the answer for a question text, or an error if no answer is present.
func (k *KeyboardInteractiveAnswers) GetByQuestionText(question string) (string, error) {
	if val, ok := k.answers[question]; ok {
		return val, nil
	}
	return "", fmt.Errorf("no answer for question")
}

// NetworkConnectionHandler is an object that is used to represent the underlying network connection and the SSH
// handshake.
type NetworkConnectionHandler interface {
	// OnAuthPassword is called when a user attempts a password authentication. The implementation must always supply
	// AuthResponse and may supply error as a reason description.
	OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, password []byte) (
		AuthResponse,
		metadata.ConnectionAuthenticatedMetadata,
		error,
	)

	// OnAuthPubKey is called when a user attempts a pubkey authentication. The implementation must always supply
	// AuthResponse and may supply error as a reason description. The pubKey parameter is an SSH key in
	// the form of "ssh-rsa KEY HERE".
	OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, pubKey auth.PublicKey) (
		AuthResponse,
		metadata.ConnectionAuthenticatedMetadata,
		error,
	)

	// OnAuthKeyboardInteractive is a callback for interactive authentication. The implementer will be passed a callback
	// function that can be used to issue challenges to the user. These challenges can, but do not have to contain
	// questions.
	OnAuthKeyboardInteractive(
		meta metadata.ConnectionAuthPendingMetadata,
		challenge func(
			instruction string,
			questions KeyboardInteractiveQuestions,
		) (answers KeyboardInteractiveAnswers, err error),
	) (AuthResponse, metadata.ConnectionAuthenticatedMetadata, error)

	// OnAuthGSSAPI returns a GSSAPIServer which can perform a GSSAPI authentication.
	OnAuthGSSAPI(metadata metadata.ConnectionMetadata) auth2.GSSAPIServer

	// OnHandshakeFailed is called when the SSH handshake failed. This method is also called after an authentication
	// failure. After this method is the connection will be closed and the OnDisconnect method will be
	// called.
	OnHandshakeFailed(metadata metadata.ConnectionMetadata, reason error)

	// OnHandshakeSuccess is called when the SSH handshake was successful. It returns metadata to process
	// requests, or failureReason to indicate that a backend error has happened. In this case, the
	// metadata will be closed and OnDisconnect will be called.
	OnHandshakeSuccess(metadata.ConnectionAuthenticatedMetadata) (
		connection SSHConnectionHandler,
		meta metadata.ConnectionAuthenticatedMetadata,
		failureReason error,
	)

	// OnDisconnect is called when the network connection is closed.
	OnDisconnect()

	// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
	// for the shutdown, after which the server should abort all running connections and return as fast as
	// possible.
	OnShutdown(shutdownContext context.Context)
}

// ChannelRejection is an error type that also contains a Message and a Reason
type ChannelRejection interface {
	message2.Message

	// Reason contains the SSH-specific reason for the rejection.
	Reason() ssh.RejectionReason
}

// SessionChannel contains a set of calls to manipulate the session channel.
type SessionChannel interface {
	// Stdin returns the reader for the standard input.
	Stdin() io.Reader
	// Stdout returns the writer for the standard output.
	Stdout() io.Writer
	// Stderr returns the writer for the standard error.
	Stderr() io.Writer
	// ExitStatus sends the program exit status to the client.
	ExitStatus(code uint32)
	// ExitSignal sends a message to the client indicating that the program exited violently.
	ExitSignal(signal string, coreDumped bool, errorMessage string, languageTag string)
	// CloseWrite sends an EOF to the client indicating that no more data will be sent on stdout or stderr.
	CloseWrite() error
	// Close closes the channel for reading and writing.
	Close() error
}

// SSHConnectionHandler represents an established SSH connection that is ready to receive requests.
type SSHConnectionHandler interface {
	// OnUnsupportedGlobalRequest captures all global SSH requests and gives the implementation an opportunity to log
	//                            the request.
	//
	// requestID is an ID uniquely identifying the request within the scope connection. The same ID may appear within
	//           a channel.
	OnUnsupportedGlobalRequest(requestID uint64, requestType string, payload []byte)

	// OnUnsupportedChannel is called when a new channel is requested of an unsupported type. This gives the implementer
	//                      the ability to log unsupported channel requests.
	//
	// channelID is an ID uniquely identifying the channel within the connection.
	// channelType is the type of channel requested by the client. We only support the "session" channel type
	// extraData contains the binary extra data submitted by the client. This is usually empty.
	OnUnsupportedChannel(channelID uint64, channelType string, extraData []byte)

	// OnSessionChannel is called when a channel of the session type is requested. The implementer must either return
	//                  the channel result if the channel was successful, or failureReason to state why the channel
	//                  should be rejected.
	//
	// channelMetadata contains the metadata for the channel.
	// extraData contains the binary extra data submitted by the client. This is usually empty.
	// session contains a set of calls that can be used to manipulate the SSH session.
	OnSessionChannel(
		channelMetadata metadata.ChannelMetadata,
		extraData []byte,
		session SessionChannel,
	) (channel SessionChannelHandler, failureReason ChannelRejection)

	// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
	//            for the shutdown, after which the server should abort all running connections and return as fast as
	//            possible.
	OnShutdown(shutdownContext context.Context)
}

// ExitStatus contains the status code with which the program exited.
// See RFC 4254 section 6.10: Returning Exit Status for details. ( https://tools.ietf.org/html/rfc4254#section-6.10 )
type ExitStatus uint32

// SessionChannelHandler is a channel of the "session" type used for interactive and non-interactive sessions
type SessionChannelHandler interface {
	//region Channel request initialization

	// OnUnsupportedChannelRequest captures channel requests of unsupported types.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// RequestType contains the SSH request type.
	// payload is the binary payload.
	OnUnsupportedChannelRequest(
		requestID uint64,
		requestType string,
		payload []byte,
	)

	// OnFailedDecodeChannelRequest is called when a supported channel request was received, but the payload could not
	//                              be decoded.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// RequestType contains the SSH request type.
	// payload is the binary payload.
	// reason is the reason why the decoding failed.
	OnFailedDecodeChannelRequest(
		requestID uint64,
		requestType string,
		payload []byte,
		reason error,
	)

	//endregion

	//region Requests before program execution

	// OnEnvRequest is called when the client requests an environment variable to be set. The implementation can return
	//              an error to reject the request.
	OnEnvRequest(
		requestID uint64,
		name string,
		value string,
	) error

	// OnPtyRequest is called when the client requests an interactive terminal to be allocated. The implementation can
	//              return an error to reject the request.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// Term is the terminal Name. This is usually set in the TERM environment variable.
	// Columns is the number of Columns in the terminal.
	// Rows is the number of Rows in the terminal.
	// Width is the Width of the terminal in pixels.
	// Height is the Height of a terminal in pixels.
	// ModeList are the encoded terminal modes the client desires. See RFC4254 section 8 and RFC8160 for details.
	OnPtyRequest(
		requestID uint64,
		term string,
		columns uint32,
		rows uint32,
		width uint32,
		height uint32,
		modeList []byte,
	) error

	//endregion

	//region Program execution

	// OnExecRequest is called when the client request a program to be executed. The implementation can return an error
	//               to reject the request. This method MUST NOT block beyond initializing the program.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// program is the Name of the program to be executed.
	OnExecRequest(
		requestID uint64,
		program string,
	) error

	// OnShell is called when the client requests a shell to be started. The implementation can return an error to
	//         reject the request. The implementation should send the IO handling into background. It should also
	//         respect the shutdown context on the Handler. This method MUST NOT block beyond initializing the shell.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// stdin is a reader for the shell or program to read the stdin.
	// stdout is a writer for the shell or program standard output.
	// stderr is a writer for the shell or program standard error.
	// writeClose closes the stdout and stderr for writing.
	// onExit is a callback to send the exit status back to the client.
	OnShell(
		requestID uint64,
	) error

	// OnSubsystem is called when the client calls a well-known Subsystem (e.g. sftp). The implementation can return an
	//             error to reject the request. The implementation should send the IO handling into background. It
	//             should also respect the shutdown context on the Handler. This method MUST NOT block beyond
	//             initializing the subsystem.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// subsystem is the name of the subsystem to be launched (e.g. sftp)
	OnSubsystem(
		requestID uint64,
		subsystem string,
	) error

	//endregion

	//region Requests during program execution

	// OnSignal is called when the client requests a Signal to be sent to the running process. The implementation can
	//          return an error to reject the request.
	OnSignal(
		requestID uint64,
		signal string,
	) error

	// OnWindow is called when the client requests requests the window size to be changed. This method may be called
	//          after a program is started. The implementation can return an error to reject the request.
	//
	// requestID is an incrementing number uniquely identifying this request within the channel.
	// Columns is the number of Columns in the terminal.
	// Rows is the number of Rows in the terminal.
	// Width is the Width of the terminal in pixels.
	// Height is the Height of a terminal in pixels.
	OnWindow(
		requestID uint64,
		columns uint32,
		rows uint32,
		width uint32,
		height uint32,
	) error

	//endregion

	//region closing the channel

	// OnClose is called when the channel is closed.
	OnClose()

	// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
	//            for the shutdown, after which the server should abort all running connections and return as fast as
	//            possible.
	OnShutdown(shutdownContext context.Context)

	//endregion
}
