package message

// ConnectionID is an opaque, globally unique identifier for a connection made to the SSH server
type ConnectionID string

// ChannelID is the ID of an SSH channel
type ChannelID *uint64

func MakeChannelID(n uint64) ChannelID {
	return &n
}

// Message is a basic element of audit logging. It contains the basic records of an interaction.
type Message struct {
	ConnectionID ConnectionID `json:"connectionId" yaml:"connectionId"` // ConnectionID is an opaque ID of the connection.
	Timestamp    int64        `json:"timestamp" yaml:"timestamp"`       // Timestamp is a nanosecond timestamp when the message was created.
	MessageType  Type         `json:"type" yaml:"type"`                 // Type of the Payload object.
	Payload      Payload      `json:"payload" yaml:"payload"`           // Payload is always a pointer to a payload object.
	ChannelID    ChannelID    `json:"channelId" yaml:"channelId"`       // ChannelID is an identifier for an SSH channel, if applicable. -1 otherwise.
}

// GetExtendedMessage returns a message with the added human-readable typeName field.
func (m Message) GetExtendedMessage() ExtendedMessage {
	return ExtendedMessage{
		m.ConnectionID,
		m.Timestamp,
		m.MessageType,
		m.MessageType.ID(),
		m.MessageType.Name(),
		m.Payload,
		m.ChannelID,
	}
}

// Payload is an interface that makes sure all payloads with Message have a method to compare them.
type Payload interface {
	// Equals compares if the current payload is identical to the provided other payload.
	Equals(payload Payload) bool
}

// Equals is a method to compare two messages with each other.
func (m Message) Equals(other Message) bool {
	if m.ConnectionID != other.ConnectionID {
		return false
	}
	if m.Timestamp != other.Timestamp {
		return false
	}
	if m.MessageType != other.MessageType {
		return false
	}
	if m.ChannelID != other.ChannelID {
		if m.ChannelID != nil && other.ChannelID != nil {
			if !(*m.ChannelID == *other.ChannelID) {
				return false
			}
		} else {
			return false
		}
	}

	if m.Payload == nil && other.Payload != nil {
		return false
	}
	if m.Payload != nil && other.Payload == nil {
		return false
	}
	if m.Payload != nil {
		return m.Payload.Equals(other.Payload)
	}
	return true
}
