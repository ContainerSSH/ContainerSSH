package sshserver

import (
	"context"

	messageCodes "github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
	"golang.org/x/crypto/ssh"
)

// AbstractSSHConnectionHandler is an empty implementation of the SSHConnectionHandler providing default methods.
type AbstractSSHConnectionHandler struct {
}

// OnUnsupportedGlobalRequest captures all global SSH requests and gives the implementation an opportunity to log
//                            the request.
//
// requestID is an ID uniquely identifying the request within the scope connection. The same ID may appear within
//           a channel.
func (a *AbstractSSHConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {}

// OnUnsupportedChannel is called when a new channel is requested of an unsupported type. This gives the implementer
//                      the ability to log unsupported channel requests.
//
// channelID is an ID uniquely identifying the channel within the connection.
// channelType is the type of channel requested by the client. We only support the "session" channel type
// extraData contains the binary extra data submitted by the client. This is usually empty.
func (a *AbstractSSHConnectionHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {}

// OnSessionChannel is called when a channel of the session type is requested. The implementer must either return
//                  the channel result if the channel was successful, or failureReason to state why the channel
//                  should be rejected.
func (a *AbstractSSHConnectionHandler) OnSessionChannel(_ metadata.ChannelMetadata, _ []byte, _ SessionChannel) (
	channel SessionChannelHandler, failureReason ChannelRejection,
) {
	return nil, NewChannelRejection(
		ssh.UnknownChannelType,
		messageCodes.ESSHNotImplemented,
		"Cannot open session channel.",
		"Session channels are currently not implemented",
	)
}

// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
//            for the shutdown, after which the server should abort all running connections and return as fast as
//            possible.
func (a *AbstractSSHConnectionHandler) OnShutdown(_ context.Context) {}
