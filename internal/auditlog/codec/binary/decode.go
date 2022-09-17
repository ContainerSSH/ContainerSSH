package binary

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/fxamacker/cbor"
	"github.com/mitchellh/mapstructure"
	"go.containerssh.io/libcontainerssh/auditlog/message"
	"go.containerssh.io/libcontainerssh/internal/auditlog/codec"
)

// NewDecoder Creates a decoder for the CBOR+GZIP audit log format.
func NewDecoder() codec.Decoder {
	return &decoder{}
}

type decoder struct {
}

func (d *decoder) Decode(reader io.Reader) (<-chan message.Message, <-chan error) {
	result := make(chan message.Message)
	errors := make(chan error)

	if err := readHeader(reader, CurrentVersion); err != nil {
		go func() {
			errors <- err
			close(result)
			close(errors)
		}()
		return result, errors
	}

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		go func() {
			errors <- fmt.Errorf("failed to open gzip stream (%w)", err)
			close(result)
			close(errors)
		}()
		return result, errors
	}

	cborReader := cbor.NewDecoder(gzipReader)

	var messages []decodedMessage

	go func() {
		if err = cborReader.Decode(&messages); err != nil {
			errors <- fmt.Errorf("failed to decode messages (%w)", err)
			close(result)
			close(errors)
			return
		}

		for _, v := range messages {
			decodedMessage, err := decodeMessage(v)
			if err != nil {
				errors <- err
			} else {
				result <- *decodedMessage
			}
		}
		close(result)
		close(errors)
	}()
	return result, errors
}

type decodedMessage struct {
	// ConnectionID is an opaque ID of the connection
	ConnectionID message.ConnectionID `json:"connectionId" yaml:"connectionId"`
	// Timestamp is a nanosecond timestamp when the message was created
	Timestamp int64 `json:"timestamp" yaml:"timestamp"`
	// Type of the Payload object
	MessageType message.Type `json:"type" yaml:"type"`
	// Payload is always a pointer to a payload object.
	Payload map[string]interface{} `json:"payload" yaml:"payload"`
	// ChannelID is an identifier for an SSH channel, if applicable. -1 otherwise.
	ChannelID message.ChannelID `json:"channelId" yaml:"channelId"`
}

func decodeMessage(v decodedMessage) (*message.Message, error) {
	payload, err := v.MessageType.Payload()
	if err != nil {
		return nil, err
	}

	if payload != nil {
		if err := mapstructure.Decode(v.Payload, &payload); err != nil {
			return nil, err
		}
	}
	return &message.Message{
		ConnectionID: v.ConnectionID,
		Timestamp:    v.Timestamp,
		MessageType:  v.MessageType,
		Payload:      payload,
		ChannelID:    v.ChannelID,
	}, nil
}
