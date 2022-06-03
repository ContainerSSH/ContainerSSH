package sshserver

import (
	"context"
	"fmt"
)

// AbstractSessionChannelHandler is an abstract implementation of SessionChannelHandler providing default
// implementations.
type AbstractSessionChannelHandler struct {
}

// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
//            for the shutdown, after which the server should abort all running connections and return as fast as
//            possible.
func (a *AbstractSessionChannelHandler) OnShutdown(_ context.Context) {}

// OnClose is called when the channel is closed.
func (a *AbstractSessionChannelHandler) OnClose() {}

// OnUnsupportedChannelRequest captures channel requests of unsupported types.
//
// requestID is an incrementing number uniquely identifying this request within the channel.
// RequestType contains the SSH request type.
// payload is the binary payload.
func (a *AbstractSessionChannelHandler) OnUnsupportedChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
) {
}

// OnFailedDecodeChannelRequest is called when a supported channel request was received, but the payload could not
//                              be decoded.
//
// requestID is an incrementing number uniquely identifying this request within the channel.
// RequestType contains the SSH request type.
// payload is the binary payload.
// reason is the reason why the decoding failed.
func (a *AbstractSessionChannelHandler) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {
}

// OnEnvRequest is called when the client requests an environment variable to be set. The implementation can return
//              an error to reject the request.
func (a *AbstractSessionChannelHandler) OnEnvRequest(
	_ uint64,
	_ string,
	_ string,
) error {
	return fmt.Errorf("not supported")
}

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
func (a *AbstractSessionChannelHandler) OnPtyRequest(
	_ uint64,
	_ string,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	return fmt.Errorf("not supported")
}

// OnExecRequest is called when the client request a program to be executed. The implementation can return an error
//               to reject the request. This method MUST NOT block beyond initializing the program.
func (a *AbstractSessionChannelHandler) OnExecRequest(
	_ uint64,
	_ string,
) error {
	return fmt.Errorf("not supported")
}

// OnShell is called when the client requests a shell to be started. The implementation can return an error to
//         reject the request. The implementation should send the IO handling into background. It should also
//         respect the shutdown context on the Handler. This method MUST NOT block beyond initializing the shell.
func (a *AbstractSessionChannelHandler) OnShell(
	_ uint64,
) error {
	return fmt.Errorf("not supported")
}

// OnSubsystem is called when the client calls a well-known Subsystem (e.g. sftp). The implementation can return an
//             error to reject the request. The implementation should send the IO handling into background. It
//             should also respect the shutdown context on the Handler. This method MUST NOT block beyond
//             initializing the subsystem.
func (a *AbstractSessionChannelHandler) OnSubsystem(
	_ uint64,
	_ string,
) error {
	return fmt.Errorf("not supported")
}

//endregion

//region Requests during program execution

// OnSignal is called when the client requests a Signal to be sent to the running process. The implementation can
//          return an error to reject the request.
func (a *AbstractSessionChannelHandler) OnSignal(
	_ uint64,
	_ string,
) error {
	return fmt.Errorf("not supported")
}

// OnWindow is called when the client requests requests the window size to be changed. This method may be called
//          after a program is started. The implementation can return an error to reject the request.
//
// requestID is an incrementing number uniquely identifying this request within the channel.
// Columns is the number of Columns in the terminal.
// Rows is the number of Rows in the terminal.
// Width is the Width of the terminal in pixels.
// Height is the Height of a terminal in pixels.
func (a *AbstractSessionChannelHandler) OnWindow(
	_ uint64,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
) error {
	return fmt.Errorf("not supported")
}

// OnX11Request is called when the client requests the forwarding of X11 connections from the container to the client.
// This method may be called after a program is started. The implementation can return an error to reject the request.
//
// requestid is an incrementing number uniquely identifying the request within the channel.
// singleConnection is a flag determining whether only one or multiple connections should be forwarded
// protocol is the authentication protocol for the X11 connections
// cookie is the authentication cookie for the X11 connections
// screen is the X11 screen number
// reverseHandler is a callback interface to signal when new connections are made
func (s *AbstractSessionChannelHandler) OnX11Request(

	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler ReverseForward,
) error {
	return fmt.Errorf("not supported")
}