package agentprotocol

import (
	"fmt"
	"io"
	"sync"
	"time"

    log "go.containerssh.io/libcontainerssh/log"
    message "go.containerssh.io/libcontainerssh/message"
	"github.com/fxamacker/cbor/v2"
)

const (
	CONNECTION_STATE_WAITINIT = iota
	CONNECTION_STATE_STARTED
	CONNECTION_STATE_WAITCLOSE
	CONNECTION_STATE_CLOSED
)

type Connection struct {
	logger        log.Logger
	lock          sync.Mutex
	state         int
	initiator     bool
	stateCond     *sync.Cond
	id            uint64
	details       NewConnectionPayload
	bufferReader  *io.PipeReader
	bufferWriter  *io.PipeWriter
	ctx           *ForwardCtx
	closeCallback func() error
}

func (c *Connection) Read(p []byte) (n int, err error) {
	return c.bufferReader.Read(p)
}

func (c *Connection) Write(data []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
L:
	for {
		switch c.state {
		case CONNECTION_STATE_WAITINIT:
			c.stateCond.Wait()
			continue
		case CONNECTION_STATE_STARTED:
			break L
		case CONNECTION_STATE_WAITCLOSE:
			fallthrough
		case CONNECTION_STATE_CLOSED:
			_ = c.bufferWriter.Close()
			return 0, fmt.Errorf("connection closed")
		default:
			return 0, fmt.Errorf("unknown connection state %d", c.state)
		}
	}

	packet := Packet{
		Type:         PACKET_DATA,
		ConnectionId: c.id,
		Payload:      data,
	}
	err = c.ctx.writePacket(&packet)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.MSSHConnected,
			"Error writing packet",
		))
		return 0, err
	}
	return len(data), nil
}

func (c *Connection) Close() error {
	c.lock.Lock()

	switch c.state {
	case CONNECTION_STATE_WAITINIT:
		fallthrough
	case CONNECTION_STATE_STARTED:
		c.state = CONNECTION_STATE_WAITCLOSE
		c.stateCond.Broadcast()
		c.lock.Unlock()
		packet := Packet{
			Type:         PACKET_CLOSE_CONNECTION,
			ConnectionId: c.id,
		}
		return c.ctx.writePacket(&packet)
	case CONNECTION_STATE_WAITCLOSE:
		c.lock.Unlock()
		return nil
	case CONNECTION_STATE_CLOSED:
		c.lock.Unlock()
		return nil
	}
	return fmt.Errorf("unknown state")
}

func (c *Connection) CloseImm() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.state != CONNECTION_STATE_WAITINIT && c.state != CONNECTION_STATE_STARTED && c.state != CONNECTION_STATE_WAITCLOSE {
		return fmt.Errorf("unclosable state")
	}
	c.state = CONNECTION_STATE_CLOSED
	if c.closeCallback != nil {
		_ = c.closeCallback()
	}
	_ = c.bufferWriter.Close()
	_ = c.bufferReader.Close()
	c.stateCond.Broadcast()
	c.ctx.waitGroup.Done()
	return nil
}

func (c *Connection) Accept() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.initiator {
		return fmt.Errorf("cannot accept connection that was initiated locally")
	}
	if c.state != CONNECTION_STATE_WAITINIT {
		return fmt.Errorf("invalid state, cannot accept connection in state %d", c.state)
	}
	c.state = CONNECTION_STATE_STARTED
	c.stateCond.Broadcast()
	packet := Packet{
		Type:         PACKET_SUCCESS,
		ConnectionId: c.id,
	}
	return c.ctx.writePacket(&packet)
}

func (c *Connection) Reject() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.initiator {
		return fmt.Errorf("cannot reject a connection that was initiated locally")
	}
	if c.state != CONNECTION_STATE_WAITINIT {
		return fmt.Errorf("invalid state, cannot accept connection in state %d", c.state)
	}
	c.state = CONNECTION_STATE_CLOSED
	c.stateCond.Broadcast()
	packet := Packet{
		Type:         PACKET_ERROR,
		ConnectionId: c.id,
	}
	return c.ctx.writePacket(&packet)
}

func (c *Connection) Details() NewConnectionPayload {
	return c.details
}

func (c *Connection) setState(state int) {
	c.lock.Lock()
	c.state = state
	c.stateCond.Broadcast()
	c.lock.Unlock()
}

type ForwardCtx struct {
	fromBackend       io.Reader
	toBackend         io.Writer
	logger            log.Logger
	connectionChannel chan *Connection
	stopped           bool

	connectionId uint64
	connMapMu    sync.RWMutex
	connMap      map[uint64]*Connection
	encoderMu    sync.Mutex
	encoder      *cbor.Encoder
	decoder      *cbor.Decoder

	waitGroup sync.WaitGroup
}

func (c *ForwardCtx) writePacket(packet *Packet) error {
	c.encoderMu.Lock()
	err := c.encoder.Encode(&packet)
	c.encoderMu.Unlock()
	return err
}

func (c *ForwardCtx) handleData(packet *Packet) {
	c.connMapMu.RLock()
	conn, ok := c.connMap[packet.ConnectionId]
	c.connMapMu.RUnlock()
	if !ok {
		c.logger.Info(
			message.NewMessage(
				message.EAgentUnknownConnection,
				"Received data packet with unknown connection id %d",
				packet.ConnectionId,
			),
		)
		return
	}
	conn.lock.Lock()
	defer conn.lock.Unlock()
	if conn.state != CONNECTION_STATE_STARTED {
		c.logger.Info(
			message.NewMessage(
				message.EAgentConnectionInvalidState,
				"Received data packet for a connection in a non-started state",
			),
		)
		return
	}
	nByte, err := conn.bufferWriter.Write(packet.Payload)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.MSSHConnected,
			"Error handling data packet",
		))
		return
	}
	if nByte != len(packet.Payload) {
		c.logger.Warning(
			message.NewMessage(
				message.EAgentWriteFailed,
				"Failed to write connection packet to agent",
			),
		)
		return
	}
}

func (c *ForwardCtx) handleClose(packet *Packet) {
	c.connMapMu.Lock()
	conn, ok := c.connMap[packet.ConnectionId]
	if !ok {
		c.logger.Info(
			message.NewMessage(
				message.EAgentUnknownConnection,
				"Received close packet with unknown connection id %d",
				packet.ConnectionId,
			),
		)
		return
	}
	c.connMapMu.Unlock()
	err := conn.CloseImm()
	retPacket := Packet{
		Type:         PACKET_SUCCESS,
		ConnectionId: conn.id,
	}
	if err != nil {
		retPacket.Type = PACKET_ERROR
	}
	_ = c.writePacket(&retPacket)
}

func (c *ForwardCtx) handleSuccess(packet *Packet) {
	c.connMapMu.Lock()
	defer c.connMapMu.Unlock()
	conn, ok := c.connMap[packet.ConnectionId]
	if !ok {
		c.logger.Info(
			message.NewMessage(
				message.EAgentUnknownConnection,
				"Received success packet with unknown connection id %d",
				packet.ConnectionId,
			),
		)
		return
	}

	switch conn.state {
	case CONNECTION_STATE_WAITINIT:
		conn.setState(CONNECTION_STATE_STARTED)
	case CONNECTION_STATE_WAITCLOSE:
		_ = conn.CloseImm()
	default:
		c.logger.Warning(
			message.NewMessage(
				message.EAgentConnectionInvalidState,
				"Received success packet for agent connection in non-wait state",
			),
		)
	}
}

func (c *ForwardCtx) handleError(packet *Packet) {
	c.connMapMu.Lock()
	defer c.connMapMu.Unlock()
	conn, ok := c.connMap[packet.ConnectionId]
	if !ok {
		c.logger.Info(
			message.NewMessage(
				message.EAgentUnknownConnection,
				"Received error packet with unknown connection id %d",
				packet.ConnectionId,
			),
		)
		return
	}

	c.logger.Info(
		message.NewMessage(
			message.MAgentRemoteError,
			"Received error packet for connection %d from remote",
			packet.ConnectionId,
		),
	)

	_ = conn.CloseImm()
}

func (c *ForwardCtx) handleNewConnection(packet *Packet) {
	newConnectionPacket, err := c.unmarshalNewConnection(packet.Payload)
	if err != nil {
		c.logger.Error("Error unmarshalling new connection payload", err)
		return
	}
	pipeReader, pipeWriter := io.Pipe()
	connection := Connection{
		state:        CONNECTION_STATE_WAITINIT,
		id:           packet.ConnectionId,
		details:      newConnectionPacket,
		bufferReader: pipeReader,
		bufferWriter: pipeWriter,
		ctx:          c,
		logger:       c.logger,
	}
	connection.stateCond = sync.NewCond(&connection.lock)
	c.connMapMu.Lock()
	if _, ok := c.connMap[packet.ConnectionId]; ok {
		c.logger.Warning("Remote tried to open connection with re-used connectionId")
		// Cannot send reject here, might interfere with other connection ?
		c.connMapMu.Unlock()
		return
	}
	if packet.ConnectionId <= c.connectionId {
		c.logger.Warning("Suspicious connection, id <= prev")
		// Can't send reject here either
		c.connMapMu.Unlock()
		return
	}
	if packet.ConnectionId != c.connectionId+1 {
		c.logger.Warning("Suspicious connection, id not prev + 1")
	}

	c.connectionId = packet.ConnectionId
	c.connMap[packet.ConnectionId] = &connection
	c.waitGroup.Add(1)
	c.connMapMu.Unlock()

	if c.stopped {
		c.logger.Warning("Client tried opening a connection after stopping")
		_ = connection.Reject()
		return
	}

	c.connectionChannel <- &connection
}

func (c *ForwardCtx) handleBackend() {
	for {
		packet := Packet{}
		err := c.decoder.Decode(&packet)
		if err != nil {
			c.logger.Error(message.Wrap(
				err,
				message.MSSHConnected,
				"Error decoding packet from backend",
			))
			return
		}
		switch packet.Type {
		case PACKET_DATA:
			c.handleData(&packet)
		case PACKET_CLOSE_CONNECTION:
			c.handleClose(&packet)
		case PACKET_SUCCESS:
			c.handleSuccess(&packet)
		case PACKET_ERROR:
			c.handleError(&packet)
		case PACKET_NEW_CONNECTION:
			c.handleNewConnection(&packet)
		case PACKET_NO_MORE_CONNECTIONS:
			if !c.stopped {
				c.stopped = true
				close(c.connectionChannel)
			}
		default:
			c.logger.Warning(
				message.NewMessage(
					message.EAgentUnknownPacket,
					"Received unknown packet type %d from agent",
					packet.Type,
				),
			)
		}
	}
}

func (c *ForwardCtx) unmarshalSetup(payload []byte) (SetupPacket, error) {
	packet := SetupPacket{}
	err := cbor.Unmarshal(payload, &packet)
	if err != nil {
		return packet, err
	}
	return packet, nil
}

func (c *ForwardCtx) unmarshalNewConnection(payload []byte) (NewConnectionPayload, error) {
	packet := NewConnectionPayload{}
	err := cbor.Unmarshal(payload, &packet)
	if err != nil {
		return packet, err
	}
	return packet, nil
}

func (c *ForwardCtx) NewConnectionTCP(
	connectedAddress string,
	connectedPort uint32,
	origAddress string,
	origPort uint32,
	closeFunc func() error,
) (io.ReadWriteCloser, error) {
	return c.newConnection(
		PROTOCOL_TCP,
		connectedAddress,
		connectedPort,
		origAddress,
		origPort,
		closeFunc,
	)
}

func (c *ForwardCtx) NewConnectionUnix(
	path string,
	closeFunc func() error,
) (io.ReadWriteCloser, error) {
	return c.newConnection(
		PROTOCOL_UNIX,
		path,
		0,
		"",
		0,
		closeFunc,
	)
}

func (c *ForwardCtx) newConnection(
	protocol string,
	connectedAddress string,
	connectedPort uint32,
	origAddress string,
	origPort uint32,
	closeFunc func() error,
) (io.ReadWriteCloser, error) {
	connInfo := NewConnectionPayload{
		Protocol:          protocol,
		ConnectedAddress:  connectedAddress,
		ConnectedPort:     connectedPort,
		OriginatorAddress: origAddress,
		OriginatorPort:    origPort,
	}
	marInfo, err := cbor.Marshal(&connInfo)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.MSSHConnected,
			"Error marshalling new connection payload",
		))
		return nil, err
	}

	bufferReader, bufferWriter := io.Pipe()
	conn := Connection{
		state:         CONNECTION_STATE_WAITINIT,
		initiator:     true,
		bufferReader:  bufferReader,
		bufferWriter:  bufferWriter,
		ctx:           c,
		logger:        c.logger,
		closeCallback: closeFunc,
	}
	conn.stateCond = sync.NewCond(&conn.lock)

	c.connMapMu.Lock()
	c.connectionId += 1
	conn.id = c.connectionId
	if _, ok := c.connMap[conn.id]; ok {
		return nil, fmt.Errorf("Connection id already exists, something went terribly wrong")
	}
	c.connMap[conn.id] = &conn
	c.waitGroup.Add(1)
	c.connMapMu.Unlock()
	err = c.writePacket(&Packet{
		Type:         PACKET_NEW_CONNECTION,
		ConnectionId: conn.id,
		Payload:      marInfo,
	})
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.MSSHConnected,
			"Error writing new connection packet",
		))
		return nil, err
	}

	return &conn, nil
}

func (c *ForwardCtx) init() {
	c.connMap = make(map[uint64]*Connection)
	c.connectionChannel = make(chan *Connection)

	c.encoder = cbor.NewEncoder(c.toBackend)
	c.decoder = cbor.NewDecoder(c.fromBackend)
}

func (c *ForwardCtx) StartClient() (connectionType uint32, setupPacket SetupPacket, connChan chan *Connection, err error) {
	c.init()

	packet := Packet{}
	err = c.decoder.Decode(&packet)
	if err != nil {
		c.logger.Warning("Failed to decode packet")
		return 0, SetupPacket{}, nil, err
	}
	if packet.Type != PACKET_SETUP {
		c.logger.Warning(
			message.NewMessage(
				message.EAgentPacketInvalid,
				"Received packet type %d when expecting startup packet from agent",
				packet.Type,
			),
		)
		return 0, SetupPacket{}, nil, fmt.Errorf("invalid packet type, expecting PACKET_SETUP")
	}
	setup, err := c.unmarshalSetup(packet.Payload)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.EAgentDecodingFailed,
			"Error unmarshalling setup packet",
		))
		return 0, setup, nil, err
	}

	success := Packet{
		Type: PACKET_SUCCESS,
	}
	err = c.writePacket(&success)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.EAgentWriteFailed,
			"Error writing success packet",
		))
		return 0, setup, nil, err
	}

	go c.handleBackend()

	return setup.ConnectionType, setup, c.connectionChannel, nil
}

func (c *ForwardCtx) StartServerForward() (chan *Connection, error) {
	c.init()

	setupPacket := SetupPacket{
		ConnectionType: CONNECTION_TYPE_PORT_DIAL,
	}
	mar, err := cbor.Marshal(&setupPacket)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.EAgentDecodingFailed,
			"Error marshalling setup packet",
		))
		return nil, err
	}

	packet := Packet{
		Type:    PACKET_SETUP,
		Payload: mar,
	}
	err = c.writePacket(&packet)
	if err != nil {
		return nil, err
	}

	resp := Packet{}
	err = c.decoder.Decode(&resp)
	if err != nil {
		return nil, err
	}
	if resp.Type == PACKET_ERROR {
		return nil, fmt.Errorf("received error packet from client")
	} else if resp.Type != PACKET_SUCCESS {
		return nil, fmt.Errorf("received invalid packet from client")
	}

	go c.handleBackend()

	return c.connectionChannel, nil
}

func (c *ForwardCtx) startReverseForwardingClient(setupPacket SetupPacket) (chan *Connection, error) {
	c.init()

	mar, err := cbor.Marshal(&setupPacket)
	if err != nil {
		c.logger.Error(message.Wrap(
			err,
			message.EAgentDecodingFailed,
			"Error marshalling setup packet",
		))
		return nil, err
	}

	packet := Packet{
		Type:    PACKET_SETUP,
		Payload: mar,
	}
	err = c.writePacket(&packet)
	if err != nil {
		return nil, err
	}

	resp := Packet{}
	err = c.decoder.Decode(&resp)
	if err != nil {
		return nil, err
	}
	if resp.Type == PACKET_ERROR {
		return nil, fmt.Errorf("received error packet from client")
	} else if resp.Type != PACKET_SUCCESS {
		return nil, fmt.Errorf("received invalid packet from client")
	}

	go c.handleBackend()

	return c.connectionChannel, nil
}

func (c *ForwardCtx) StartX11ForwardClient(singleConnection bool, screen string, authProtocol string, authCookie string) (chan *Connection, error) {
	setupPacket := SetupPacket{
		ConnectionType:   CONNECTION_TYPE_X11,
		Protocol:         "tcp",
		SingleConnection: singleConnection,
		Screen:           screen,
		AuthProtocol:     authProtocol,
		AuthCookie:       authCookie,
	}

	return c.startReverseForwardingClient(setupPacket)
}

func (c *ForwardCtx) StartReverseForwardClient(bindHost string, bindPort uint32, singleConnection bool) (chan *Connection, error) {
	setupPacket := SetupPacket{
		ConnectionType:   CONNECTION_TYPE_PORT_FORWARD,
		BindHost:         bindHost,
		BindPort:         bindPort,
		Protocol:         "tcp",
		SingleConnection: singleConnection,
	}

	return c.startReverseForwardingClient(setupPacket)
}

func (c *ForwardCtx) StartReverseForwardClientUnix(path string, singleConnection bool) (chan *Connection, error) {
	setupPacket := SetupPacket{
		ConnectionType:   CONNECTION_TYPE_PORT_FORWARD,
		BindHost:         path,
		Protocol:         "unix",
		SingleConnection: singleConnection,
	}

	return c.startReverseForwardingClient(setupPacket)
}

func (c *ForwardCtx) NoMoreConnections() error {
	c.stopped = true
	close(c.connectionChannel)
	return c.writePacket(
		&Packet{
			Type: PACKET_NO_MORE_CONNECTIONS,
		},
	)
}

func (c *ForwardCtx) WaitFinish() {
	c.waitGroup.Wait()
}

func (c *ForwardCtx) Kill() {
	if !c.stopped {
		_ = c.NoMoreConnections()
	}
	for _, conn := range c.connMap {
		_ = conn.Close()
	}
	t := make(chan struct{})
	go func() {
		select {
		case <-t:
		case <-time.After(5 * time.Second):
			for _, conn := range c.connMap {
				_ = conn.CloseImm()
			}
		}
	}()
	c.WaitFinish()
}
