package format

type ChannelID int64

type Message struct {
	ConnectionID []byte      `json:"connectionId" yaml:"connectionId"`
	Timestamp    int64       `json:"timestamp" yaml:"timestamp"`
	MessageType  MessageType `json:"type" yaml:"type"`
	Payload      interface{} `json:"payload" yaml:"payload"`
	ChannelID    ChannelID
}
