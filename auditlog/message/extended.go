package message

// ExtendedMessage is a message that also contains the type in a text format.
type ExtendedMessage struct {
	ConnectionID ConnectionID `json:"connectionId" yaml:"connectionId"` // ConnectionID is an opaque ID of the connection.
	Timestamp    int64        `json:"timestamp" yaml:"timestamp"`       // Timestamp is a nanosecond timestamp when the message was created.
	MessageType  Type         `json:"type" yaml:"type"`                 // Type of the Payload object.
	TypeID       string       `json:"typeId" yaml:"typeId"`             // TypeID is a machine-readable text ID of the message type.
	TypeName     string       `json:"typeName" yaml:"typeName"`         // TypeName is the human-readable name of the message type.
	Payload      Payload      `json:"payload" yaml:"payload"`           // Payload is always a pointer to a payload object.
	ChannelID    ChannelID    `json:"channelId" yaml:"channelId"`       // ChannelID is an identifier for an SSH channel, if applicable. -1 otherwise.
}
