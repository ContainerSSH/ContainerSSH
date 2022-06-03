package asciinema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

    "go.containerssh.io/libcontainerssh/auditlog/message"
    asciinema2 "go.containerssh.io/libcontainerssh/internal/auditlog/codec/asciinema"
    "go.containerssh.io/libcontainerssh/internal/geoip/dummy"
    "go.containerssh.io/libcontainerssh/log"
	"github.com/stretchr/testify/assert"
)

type writer struct {
	data      bytes.Buffer
	startTime int64
	sourceIP  string
	username  *string
	wait      chan bool
	country   string
}

func newWriter() *writer {
	return &writer{
		wait: make(chan bool),
	}
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.data.Write(p)
}

func (w *writer) Close() error {
	w.wait <- true
	return nil
}

func (w *writer) waitForClose() {
	<-w.wait
}

func (w *writer) SetMetadata(startTime int64, sourceIP string, country string, username *string) {
	w.startTime = startTime
	w.sourceIP = sourceIP
	w.username = username
	w.country = country
}

func sendMessagesAndReturnWrittenData(
	t *testing.T,
	messages []message.Message,
) (asciinema2.Header, []asciinema2.Frame, error) {
	logger := log.NewTestLogger(t)
	geoIPProvider := dummy.New()
	encoder := asciinema2.NewEncoder(logger, geoIPProvider)
	msgChannel := make(chan message.Message)
	writer := newWriter()
	go func() {
		if err := encoder.Encode(msgChannel, writer); err != nil {
			assert.Fail(t, "failed to encode messages", err)
		}
	}()

	for _, msg := range messages {
		msgChannel <- msg
	}
	close(msgChannel)
	writer.waitForClose()

	lines := strings.Split(writer.data.String(), "\n")

	header := asciinema2.Header{}
	var frames []asciinema2.Frame

	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		return header, frames, fmt.Errorf("failed to unmarshal header line (%w)", err)
	}

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		frame := asciinema2.Frame{}
		if err := json.Unmarshal([]byte(line), &frame); err != nil {
			return header, frames, fmt.Errorf("failed to unmarshal frame (%w)", err)
		}
		frames = append(frames, frame)
	}
	return header, frames, nil
}

func TestHeader(t *testing.T) {
	header, frames, err := sendMessagesAndReturnWrittenData(t, []message.Message{
		{
			ConnectionID: "0123456789ABCDEF",
			Timestamp:    0,
			MessageType:  message.TypeConnect,
			Payload: message.PayloadConnect{
				RemoteAddr: "127.0.0.1",
			},
			ChannelID: nil,
		},
		{
			ConnectionID: "0123456789ABCDEF",
			Timestamp:    0,
			MessageType:  message.TypeDisconnect,
			Payload:      nil,
			ChannelID:    nil,
		},
	})
	if err != nil {
		assert.Fail(t, "failed to process messages", err)
		return
	}

	assert.Equal(t, 0, len(frames))

	assert.Equal(t, uint(2), header.Version)
	assert.Equal(t, 0, header.Timestamp)
	assert.True(t, header.Height > 0)
	assert.True(t, header.Width > 0)
}

var fullOutputTestMessages = []message.Message{
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    0,
		MessageType:  message.TypeConnect,
		Payload: message.PayloadConnect{
			RemoteAddr: "127.0.0.1",
		},
		ChannelID: nil,
	},
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    int64(time.Second),
		MessageType:  message.TypeNewChannel,
		Payload: message.PayloadNewChannel{
			ChannelType: "session",
		},
		ChannelID: nil,
	},
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    int64(2 * time.Second),
		MessageType:  message.TypeNewChannelSuccessful,
		Payload: message.PayloadNewChannelSuccessful{
			ChannelType: "session",
		},
		ChannelID: message.MakeChannelID(0),
	},
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    int64(3 * time.Second),
		MessageType:  message.TypeChannelRequestShell,
		Payload:      message.PayloadChannelRequestShell{},
		ChannelID:    message.MakeChannelID(0),
	},
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    int64(4 * time.Second),
		MessageType:  message.TypeIO,
		Payload: message.PayloadIO{
			Stream: message.StreamStdout,
			Data:   []byte("Hello world!"),
		},
		ChannelID: message.MakeChannelID(0),
	},
	{
		ConnectionID: "0123456789ABCDEF",
		Timestamp:    int64(5 * time.Second),
		MessageType:  message.TypeDisconnect,
		Payload:      nil,
		ChannelID:    nil,
	},
}

func TestOutput(t *testing.T) {
	header, frames, err := sendMessagesAndReturnWrittenData(t, fullOutputTestMessages)
	if err != nil {
		assert.Fail(t, "failed to process messages", err)
		return
	}

	assert.Equal(t, uint(2), header.Version)
	assert.Equal(t, int(fullOutputTestMessages[0].Timestamp/1000000000), header.Timestamp)
	assert.True(t, header.Height > 0)
	assert.True(t, header.Width > 0)

	assert.Equal(t, 1, len(frames))

	assert.Equal(t, float64(fullOutputTestMessages[4].Timestamp)/1000000000, frames[0].Time)
	assert.Equal(t, asciinema2.EventTypeOutput, frames[0].EventType)
	assert.Equal(t, string(fullOutputTestMessages[4].Payload.(message.PayloadIO).Data), frames[0].Data)
}
