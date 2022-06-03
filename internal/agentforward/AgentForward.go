package agentforward

import (
	"io"

	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

// AgentForward is a network connection forwarding interface that uses the ContainerSSH Agent protocol
type AgentForward interface {
	// NewX11Forwarding initializes the X11 forwarding mode of the agent
	//
	// setupAgentCallback is a function that should start the agent on the desired target if it's called. It should return an interface to the stdin and stdout of a new instance of the agent.
	// logger is the logging interface to be used
	// singleConnection determines whether to stop listening for connections after the first one is accepted
	// proto is the X11 authentication protocol to be used
	// cookie is the X11 authentication cookie
	// screen is the X11 screen number
	// reverseHandler is an interface that notifies the caller of new connections
	NewX11Forwarding(
		setupAgentCallback func() (io.Reader, io.Writer, error),
		logger log.Logger,
		singleConnection bool,
		proto string,
		cookie string,
		screen uint32,
		reverseHandler sshserver.ReverseForward,
	) error

	// NewTCPReverseForwarding initializes the TCP reverse forwarding mode of the agent
	//
	// setupAgentCallback is a function that should start the agent on the desired target if it's called. It should return an interface to the stdin and stdout of a new instance of the agent.
	// logger is the logging interface to be used
	// bindHost is the hostname or address to listen on for connections
	// bindPort is the port to listen on for connections
	// reverseHandler is an interface that notifies the caller of new connections
	NewTCPReverseForwarding(
		setupAgentCallback func() (io.Reader, io.Writer, error),
		logger log.Logger,
		bindHost string,
		bindPort uint32,
		reverseHandler sshserver.ReverseForward,
	) error

	// NewTCPReverseForwarding initializes the socket reverse forwarding mode of the agent
	//
	// setupAgentCallback is a function that should start the agent on the desired target if it's called. It should return an interface to the stdin and stdout of a new instance of the agent.
	// logger is the logging interface to be used
	// path is path to the unix socket that will be listened on
	// reverseHandler is an interface that notifies the caller of new connections
	NewUnixReverseForwarding(
		setupAgentCallback func() (io.Reader, io.Writer, error),
		logger log.Logger,
		path string,
		reverseHandler sshserver.ReverseForward,
	) error

	// CancelTCPForwarding stops accepting new connections for the specified TCP forwarding. Existing connections are left intact
	//
	// bindHost is the host that is being forwarded
	// bindPort is the port that is being forwarded
	CancelTCPForwarding(
		bindHost string,
		bindPort uint32,
	) error

	// CancelStreamLocalForwarding stops accepting new connections for the specified socket forwarding. Existing connections are left intact
	//
	// path is the path to the unix socket that is being forwarded
	CancelStreamLocalForwarding(
		path string,
	) error

	// CloseX11Forwarding stops accepting new connections for the specified X11 forwarding. Existing connections are left intact
	CloseX11Forwarding() error

	// NewForwardTCP start a new tcp forwarding connection (from the client to the agent)
	//
	// setupAgentCallback is a function that should start the agent on the desired target if it's called. It should return an interface to the stdin and stdout of a new instance of the agent.
	// logger is the logging interface to be used
	// hostToConnect is the hostname or IP address to connect to
	// portToConnect is the port to connect to
	// originatorHost is the hostname or IP address that is making the request
	// originatorPort is the port that the original request was received from
	NewForwardTCP(
		setupAgentCallback func() (io.Reader, io.Writer, error),
		logger log.Logger,
		hostToConnect string,
		portToConnect uint32,
		originatorHost string,
		originatorPort uint32,
	) (sshserver.ForwardChannel, error)

	// NewForwardTCP start a new unix forwarding connection (from the client to the agent)
	//
	// setupAgentCallback is a function that should start the agent on the desired target if it's called. It should return an interface to the stdin and stdout of a new instance of the agent.
	// logger is the logging interface to be used
	// path is the path to the unix socket to be connected with
	NewForwardUnix(
		setupAgentCallback func() (io.Reader, io.Writer, error),
		logger log.Logger,
		path string,
	) (sshserver.ForwardChannel, error)

	// OnShutdown kills all connections and instructs all agent instances to terminate
	OnShutdown()
}
