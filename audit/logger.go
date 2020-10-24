package audit

import (
	"crypto/rand"
	"github.com/containerssh/containerssh/audit/protocol"
	"github.com/containerssh/containerssh/config"
	"io"
	"sync"
	"time"
)



type Connection struct {
	lock          *sync.Mutex
	nextChannelId protocol.ChannelId
	audit         Plugin
	connectionId  protocol.MessageConnectionID
	interceptIo   bool
}

type Channel struct {
	channelId protocol.ChannelId
	*Connection
}

func GetConnection(audit Plugin, config config.AuditConfig) (*Connection, error) {
	tmpConnId := make([]byte, 16)
	_, err := rand.Read(tmpConnId)
	var connectionId protocol.MessageConnectionID
	for i := 0; i < 16; i++ {
		connectionId[i] = tmpConnId[i]
	}
	return &Connection{
		&sync.Mutex{},
		0,
		audit,
		connectionId,
		config.InterceptIO,
	}, err
}

func (connection *Connection) Message(messageType protocol.MessageType, payload interface{}) {
	connection.audit.Message(&protocol.Message{
		ConnectionID: connection.connectionId,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  messageType,
		ChannelId:    -1,
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

func (channel *Channel) Message(messageType protocol.MessageType, payload interface{}) {
	channel.Connection.audit.Message(&protocol.Message{
		ConnectionID: channel.Connection.connectionId,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  messageType,
		ChannelId:    channel.channelId,
		Payload:      payload,
	})
}

func (channel *Channel) InterceptIo(stdin io.Reader, stdOut io.Writer, stdErr io.Writer) (io.Reader, io.Writer, io.Writer) {
	if !channel.interceptIo {
		return stdin, stdOut, stdErr
	}

	return &interceptingReader{
			backend: stdin,
			stream:  protocol.Stream_Stdin,
			channel: channel,
		}, &interceptingWriter{
			backend: stdOut,
			stream:  protocol.Stream_StdOut,
			channel: channel,
		}, &interceptingWriter{
			backend: stdErr,
			stream:  protocol.Stream_StdErr,
			channel: channel,
		}
}

type interceptingReader struct {
	backend io.Reader
	stream protocol.Stream
	channel *Channel
}

func (i *interceptingReader) Read(p []byte) (n int, err error) {
	n, err = i.backend.Read(p)
	i.channel.Message(protocol.MessageType_IO, protocol.MessageIO{
		Stream: i.stream,
		Data:   p,
	})
	return n, err
}

type interceptingWriter struct {
	backend io.Writer
	stream protocol.Stream
	channel *Channel
}

func (i *interceptingWriter) Write(p []byte) (n int, err error) {
	i.channel.Message(protocol.MessageType_IO, protocol.MessageIO{
		Stream: i.stream,
		Data: p,
	})
	n, err = i.backend.Write(p)
	return n, err
}
