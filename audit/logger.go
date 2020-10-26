package audit

import (
	"crypto/rand"
	"github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/config"
	"io"
	"sync"
	"time"
)

type Connection struct {
	lock          *sync.Mutex
	nextChannelId audit.ChannelID
	audit         Plugin
	connectionId  []byte
	Intercept     config.AuditInterceptConfig
}

type Channel struct {
	channelId audit.ChannelID
	*Connection
}

func GetConnection(audit Plugin, config config.AuditConfig) (*Connection, error) {
	connectionId := make([]byte, 16)
	_, err := rand.Read(connectionId)
	return &Connection{
		&sync.Mutex{},
		0,
		audit,
		connectionId,
		config.Intercept,
	}, err
}

func (connection *Connection) Message(messageType audit.MessageType, payload interface{}) {
	connection.audit.Message(audit.Message{
		ConnectionID: connection.connectionId,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  messageType,
		ChannelID:    -1,
		Payload:      payload,
	})
}

func (connection *Connection) GetChannel() *Channel {
	connection.lock.Lock()
	defer connection.lock.Unlock()
	channelId := connection.nextChannelId
	connection.nextChannelId = connection.nextChannelId + 1
	return &Channel{
		channelId:  channelId,
		Connection: connection,
	}
}

func (channel *Channel) Message(messageType audit.MessageType, payload interface{}) {
	channel.Connection.audit.Message(audit.Message{
		ConnectionID: channel.Connection.connectionId,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  messageType,
		ChannelID:    channel.channelId,
		Payload:      payload,
	})
}

func (channel *Channel) InterceptIo(stdIn io.Reader, stdOut io.Writer, stdErr io.Writer) (io.Reader, io.Writer, io.Writer) {
	if channel.Intercept.Stdin {
		stdIn = &interceptingReader{
			backend: stdIn,
			stream:  audit.Stream_Stdin,
			channel: channel,
		}
	}
	if channel.Intercept.Stdout {
		stdOut = &interceptingWriter{
			backend: stdOut,
			stream:  audit.Stream_StdOut,
			channel: channel,
		}
	}
	if channel.Intercept.Stderr {
		stdErr = &interceptingWriter{
			backend: stdErr,
			stream:  audit.Stream_StdErr,
			channel: channel,
		}
	}

	return stdIn, stdOut, stdErr
}
