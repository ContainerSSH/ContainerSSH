package binary

import (
	"compress/gzip"
	"errors"
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
	errs := make(chan error)

	version, err := readHeader(reader, CurrentVersion)
	if err != nil {
		go func() {
			errs <- err
			close(result)
			close(errs)
		}()
		return result, errs
	}

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		go func() {
			errs <- fmt.Errorf("failed to open gzip stream (%w)", err)
			close(result)
			close(errs)
		}()
		return result, errs
	}

	cborReader := cbor.NewDecoder(gzipReader)

	go func() {
		switch version {
		case 1:
			var messages []decodedMessage
			if err = cborReader.Decode(&messages); err != nil {
				errs <- fmt.Errorf("failed to decode messages (%w)", err)
				close(result)
				close(errs)
				return
			}
			for _, v := range messages {
				decodedMessage, err := decodeMessage(v)
				if err != nil {
					errs <- err
				} else {
					result <- *decodedMessage
				}
			}
		case 2:
			for {
				var msg decodedMessage
				if err = cborReader.Decode(&msg); err != nil {
					if !errors.Is(err, io.ErrUnexpectedEOF) {
						errs <- fmt.Errorf("failed to decode messages (%w)", err)
					}
					close(result)
					close(errs)
					return
				}
				decodedMessage, err := decodeMessage(msg)
				if err != nil {
					errs <- err
					break
				} else {
					result <- *decodedMessage
				}
			}
		}

		close(result)
		close(errs)
	}()
	return result, errs
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
