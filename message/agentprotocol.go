package message

// EAgentUnknownConnection indicates that the ContainerSSH Agent integration received a packet referring to an unknown connection ID
const EAgentUnknownConnection = "AGENTPROTO_UNKNOWN_CONNECTION"

// EAgentConnectionInvalidState indicates that the ContainerSSH Agent integration received a packet that is invalid for the state of the connection it refers to
const EAgentConnectionInvalidState = "AGENTPROTO_INVALID_STATE"

// EAgentWriteFailed indicates an error in the communication channel
const EAgentWriteFailed = "AGENTPROTO_WRITE_FAILED"

// EAgentUnknownPacket indicates that the ContainerSSH Agent integration received an unknown packet
const EAgentUnknownPacket = "AGENTPROTO_PACKET_UNKNOWN"

// EAgentPacketInvalid indicates that the ContainerSSH Agent integration received a packet that it wasn't expecting
const EAgentPacketInvalid = "AGENTPROTO_PACKET_INVALID"

// EAgentDecodingFailed indicates that the ContainerSSH Agent integration failed to decode a packet
const EAgentDecodingFailed = "AGENTPROTO_DECODE_FAILED"

// MAgentRemoteError indicates that the ContainerSSH Agent integration received an error message from the remote agent
const MAgentRemoteError = "AGENTPROTO_REMOTE_ERROR"

// MAgentStarting indicates that the ContainerSSH Agent is starting up.
const MAgentStarting = "AGENT_STARTING"

// MAgentSocketSetup indicates that the ContainerSSH Agent is setting up an SSH agent socket.
const MAgentSocketSetup = "AGENT_SOCKET_SETUP"

// MAgentSocketListening indicates that the ContainerSSH Agent socket is now listening for connections.
const MAgentSocketListening = "AGENT_SOCKET_LISTENING"

// MAgentConnectionAccepted indicates that the ContainerSSH Agent accepted a new connection.
const MAgentConnectionAccepted = "AGENT_CONNECTION_ACCEPTED"

// MAgentChannelClosed indicates that the ContainerSSH Agent connection channel was closed.
const MAgentChannelClosed = "AGENT_CHANNEL_CLOSED"

// MAgentDialing indicates that the ContainerSSH Agent is dialing an external connection.
const MAgentDialing = "AGENT_DIALING"

// EAgentSocketAcceptFailed indicates that the ContainerSSH Agent socket accept failed.
const EAgentSocketAcceptFailed = "AGENT_SOCKET_ACCEPT_FAILED"
