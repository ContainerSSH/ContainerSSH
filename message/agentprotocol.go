package message

// EAgentUnknownConnection indicates that the ContainerSSH Agent integration received a packet referring to an unknown connection ID
const EAgentUnknownConnection = "AGENTPROTO_UNKNOWN_CONNECTION"

// EAgentConnectionInvalidState indicates that the ContainerSSH Agent integration received a packet that is invalid for the state of the connection it refers to
const EAgentConnectionInvalidState = "AGENTPROTO_INVALID_STATE"

// EAgentWriteFailed indicates an error in the communication channel
const EAgentWriteFailed = "AGENTPROTO_WRITE_FAILED"

// EAgentUknownPacket indicates that the ContainerSSH Agent integration received an unknown packet
const EAgentUnknownPacket = "AGENTPROTO_PACKET_UNKNOWN"

// EAgentPacketInvalid indicates that the ContainerSSH Agent integration received a packet that it wasn't expecting
const EAgentPacketInvalid = "AGENTPROTO_PACKET_INVALID"

// EAgentDecodingFailed indicates that the ContainerSSH Agent integration failed to decode a packet
const EAgentDecodingFailed = "AGENTPROTO_DECODE_FAILED"

// MAgentRemoteError indicates that the ContainerSSH Agent integration received an error message from the remote agent
const MAgentRemoteError = "AGENTPROTO_REMOTE_ERROR"
