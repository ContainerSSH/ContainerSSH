package auditlog

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

    "go.containerssh.io/libcontainerssh/auditlog/message"
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/auditlog/codec"
    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
    "go.containerssh.io/libcontainerssh/internal/geoip/geoipprovider"
    "go.containerssh.io/libcontainerssh/log"
)

type loggerImplementation struct {
	intercept   config.AuditLogInterceptConfig
	encoder     codec.Encoder
	storage     storage.WritableStorage
	logger      log.Logger
	wg          *sync.WaitGroup
	geoIPLookup geoipprovider.LookupProvider
}

type loggerConnection struct {
	l *loggerImplementation

	ip             net.TCPAddr
	messageChannel chan message.Message
	connectionID   message.ConnectionID
	lock           *sync.Mutex
	closed         bool
}

func (l *loggerConnection) log(msg message.Message) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if !l.closed {
		l.messageChannel <- msg
	}
}

type loggerChannel struct {
	c *loggerConnection

	channelID message.ChannelID
}

func (l *loggerImplementation) Shutdown(shutdownContext context.Context) {
	l.wg.Wait()
	l.storage.Shutdown(shutdownContext)
}

//region Connection

func (l *loggerImplementation) OnConnect(connectionID message.ConnectionID, ip net.TCPAddr) (Connection, error) {
	name := string(connectionID)
	writer, err := l.storage.OpenWriter(name)
	if err != nil {
		return nil, err
	}
	conn := &loggerConnection{
		l:              l,
		ip:             ip,
		connectionID:   connectionID,
		messageChannel: make(chan message.Message),
		lock:           &sync.Mutex{},
	}
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		err := l.encoder.Encode(conn.messageChannel, writer)
		if err != nil {
			l.logger.Emergency(err)
		}
	}()
	conn.log(message.Message{
		ConnectionID: connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeConnect,
		Payload: message.PayloadConnect{
			RemoteAddr: ip.IP.String(),
			Country:    l.geoIPLookup.Lookup(ip.IP),
		},
		ChannelID: nil,
	})
	return conn, nil
}

func (l *loggerConnection) OnDisconnect() {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeDisconnect,
		Payload:      nil,
		ChannelID:    nil,
	})
	l.lock.Lock()
	defer l.lock.Unlock()
	close(l.messageChannel)
	l.closed = true
}

func (l *loggerConnection) OnAuthPassword(username string, password []byte) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPassword,
		Payload: message.PayloadAuthPassword{
			Username: username,
			Password: password,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPasswordSuccess(username string, password []byte) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPasswordSuccessful,
		Payload: message.PayloadAuthPassword{
			Username: username,
			Password: password,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPasswordFailed(username string, password []byte) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPasswordFailed,
		Payload: message.PayloadAuthPassword{
			Username: username,
			Password: password,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPasswordBackendError(username string, password []byte, reason string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPasswordBackendError,
		Payload: message.PayloadAuthPasswordBackendError{
			Username: username,
			Password: password,
			Reason:   reason,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPubKey(username string, pubKey string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPubKey,
		Payload: message.PayloadAuthPubKey{
			Username: username,
			Key:      pubKey,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPubKeySuccess(username string, pubKey string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPubKeySuccessful,
		Payload: message.PayloadAuthPubKey{
			Username: username,
			Key:      pubKey,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPubKeyFailed(username string, pubKey string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPubKeyFailed,
		Payload: message.PayloadAuthPubKey{
			Username: username,
			Key:      pubKey,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthPubKeyBackendError(username string, pubKey string, reason string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthPubKeyBackendError,
		Payload: message.PayloadAuthPubKeyBackendError{
			Username: username,
			Key:      pubKey,
			Reason:   reason,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthKeyboardInteractiveChallenge(
	username string,
	instruction string,
	questions []message.KeyboardInteractiveQuestion,
) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthKeyboardInteractiveChallenge,
		Payload: message.PayloadAuthKeyboardInteractiveChallenge{
			Username:    username,
			Instruction: instruction,
			Questions:   questions,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthKeyboardInteractiveAnswer(username string, answers []message.KeyboardInteractiveAnswer) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthKeyboardInteractiveAnswer,
		Payload: message.PayloadAuthKeyboardInteractiveAnswer{
			Username: username,
			Answers:  answers,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthKeyboardInteractiveFailed(username string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthKeyboardInteractiveFailed,
		Payload: message.PayloadAuthKeyboardInteractiveFailed{
			Username: username,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnAuthKeyboardInteractiveBackendError(username string, reason string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeAuthKeyboardInteractiveBackendError,
		Payload: message.PayloadAuthKeyboardInteractiveBackendError{
			Username: username,
			Reason:   reason,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnHandshakeFailed(reason string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeHandshakeFailed,
		Payload: message.PayloadHandshakeFailed{
			Reason: reason,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnHandshakeSuccessful(username string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeHandshakeSuccessful,
		Payload: message.PayloadHandshakeSuccessful{
			Username: username,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnGlobalRequestUnknown(requestType string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeGlobalRequestUnknown,
		Payload: message.PayloadGlobalRequestUnknown{
			RequestType: requestType,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnGlobalRequestDecodeFailed(requestID uint64, requestType string, payload []byte, reason error) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp: time.Now().UnixNano(),
		MessageType: message.TypeGlobalRequestDecodeFailed,
		Payload: message.PayloadGlobalRequestDecodeFailed{
			RequestID: requestID,
			RequestType: requestType,
			Payload: payload,
			Reason: reason.Error(),
		},
	})
}

func (l *loggerConnection) OnNewChannel(channelID message.ChannelID, channelType string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewChannel,
		Payload: message.PayloadNewChannel{
			ChannelType: channelType,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnNewChannelFailed(channelID message.ChannelID, channelType string, reason string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewChannelFailed,
		Payload: message.PayloadNewChannelFailed{
			ChannelType: channelType,
			Reason:      reason,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnNewChannelSuccess(channelID message.ChannelID, channelType string) Channel {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewChannelSuccessful,
		Payload: message.PayloadNewChannelSuccessful{
			ChannelType: channelType,
		},
		ChannelID: channelID,
	})
	return &loggerChannel{
		c:         l,
		channelID: channelID,
	}
}

func (l *loggerConnection) OnRequestTCPReverseForward(bindHost string, bindPort uint32) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeRequestReverseForward,
		Payload: message.PayloadRequestReverseForward{
			BindHost: bindHost,
			BindPort: bindPort,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnRequestCancelTCPReverseForward(bindHost string, bindPort uint32) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeRequestCancelReverseForward,
		Payload: message.PayloadRequestReverseForward{
			BindHost: bindHost,
			BindPort: bindPort,
		},
		ChannelID: nil,
	})
}

func (l *loggerConnection) OnTCPForwardChannel(channelID message.ChannelID, hostToConnect string, portToConnect uint32, originatorHost string, originatorPort uint32) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewForwardChannel,
		Payload: message.PayloadNewForwardChannel{
			HostToConnect:  hostToConnect,
			PortToConnect:  portToConnect,
			OriginatorHost: originatorHost,
			OriginatorPort: originatorPort,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnReverseForwardChannel(channelID message.ChannelID, connectedHost string, connectedPort uint32, originatorHost string, originatorPort uint32) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewReverseForwardChannel,
		Payload: message.PayloadNewReverseForwardChannel{
			ConnectedHost:  connectedHost,
			ConnectedPort:  connectedPort,
			OriginatorHost: originatorHost,
			OriginatorPort: originatorPort,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnReverseStreamLocalChannel(channelID message.ChannelID, path string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewReverseStreamLocalChannel,
		Payload: message.PayloadRequestStreamLocal{
			Path: path,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnReverseX11ForwardChannel(channelID message.ChannelID, originatorHost string, originatorPort uint32) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewReverseX11ForwardChannel,
		Payload: message.PayloadNewReverseX11ForwardChannel{
			OriginatorHost: originatorHost,
			OriginatorPort: originatorPort,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnDirectStreamLocal(channelID message.ChannelID, path string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeNewForwardStreamLocalChannel,
		Payload: message.PayloadRequestStreamLocal{
			Path: path,
		},
		ChannelID: channelID,
	})
}

func (l *loggerConnection) OnRequestStreamLocal(path string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeRequestStreamLocal,
		Payload: message.PayloadRequestStreamLocal{
			Path: path,
		},
	})
}

func (l *loggerConnection) OnRequestCancelStreamLocal(path string) {
	l.log(message.Message{
		ConnectionID: l.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeRequestCancelStreamLocal,
		Payload: message.PayloadRequestStreamLocal{
			Path: path,
		},
	})
}

//endregion

//region Channel

func (l *loggerChannel) OnRequestUnknown(requestID uint64, requestType string, payload []byte) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestUnknownType,
		Payload: message.PayloadChannelRequestUnknownType{
			RequestID:   requestID,
			RequestType: requestType,
			Payload:     payload,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestDecodeFailed(requestID uint64, requestType string, payload []byte, reason string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestDecodeFailed,
		Payload: message.PayloadChannelRequestDecodeFailed{
			RequestID:   requestID,
			RequestType: requestType,
			Payload:     payload,
			Reason:      reason,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestSetEnv(requestID uint64, name string, value string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestSetEnv,
		Payload: message.PayloadChannelRequestSetEnv{
			RequestID: requestID,
			Name:      name,
			Value:     value,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestExec(requestID uint64, program string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestExec,
		Payload: message.PayloadChannelRequestExec{
			RequestID: requestID,
			Program:   program,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestPty(
	requestID uint64,
	term string,
	columns uint32,
	rows uint32,
	width uint32,
	height uint32,
	modelist []byte,
) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestPty,
		Payload: message.PayloadChannelRequestPty{
			RequestID: requestID,
			Term:      term,
			Columns:   columns,
			Rows:      rows,
			Width:     width,
			Height:    height,
			ModeList:  modelist,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestX11(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestX11,
		Payload: message.PayloadChannelRequestX11{
			RequestID:        requestID,
			SingleConnection: singleConnection,
			AuthProtocol:     protocol,
			Cookie:           cookie,
			Screen:           screen,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestShell(requestID uint64) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestShell,
		Payload: message.PayloadChannelRequestShell{
			RequestID: requestID,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestSignal(requestID uint64, signal string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestSignal,
		Payload: message.PayloadChannelRequestSignal{
			RequestID: requestID,
			Signal:    signal,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestSubsystem(requestID uint64, subsystem string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestSubsystem,
		Payload: message.PayloadChannelRequestSubsystem{
			RequestID: requestID,
			Subsystem: subsystem,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnRequestWindow(requestID uint64, columns uint32, rows uint32, width uint32, height uint32) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeChannelRequestWindow,
		Payload: message.PayloadChannelRequestWindow{
			RequestID: requestID,
			Columns:   columns,
			Rows:      rows,
			Width:     width,
			Height:    height,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) io(stream message.Stream, data []byte) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeIO,
		Payload: message.PayloadIO{
			Stream: stream,
			Data:   data,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) GetForwardingProxy(forward io.ReadWriteCloser) io.ReadWriteCloser {
	if !l.c.l.intercept.Forwarding {
		return forward
	}
	return &interceptingReadWriteCloser{
		backend: forward,
		reader: interceptingReader{
			backend: forward,
			stream:  message.StreamStdin,
			channel: l,
		},
		writer: interceptingWriter{
			backend: forward,
			stream:  message.StreamStdout,
			channel: l,
		},
	}
}

func (l *loggerChannel) GetStdinProxy(stdin io.Reader) io.Reader {
	if !l.c.l.intercept.Stdin {
		return stdin
	}
	return &interceptingReader{
		backend: stdin,
		stream:  message.StreamStdin,
		channel: l,
	}
}

func (l *loggerChannel) GetStdoutProxy(stdout io.Writer) io.Writer {
	if !l.c.l.intercept.Stdout {
		return stdout
	}
	return &interceptingWriter{
		backend: stdout,
		stream:  message.StreamStdout,
		channel: l,
	}
}

func (l *loggerChannel) GetStderrProxy(stderr io.Writer) io.Writer {
	if !l.c.l.intercept.Stdout {
		return stderr
	}
	return &interceptingWriter{
		backend: stderr,
		stream:  message.StreamStderr,
		channel: l,
	}
}

func (l *loggerChannel) OnRequestFailed(requestID uint64, reason error) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeRequestFailed,
		Payload: message.PayloadRequestFailed{
			RequestID: requestID,
			Reason:    reason.Error(),
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnExit(exitStatus uint32) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeExit,
		Payload: message.PayloadExit{
			ExitStatus: exitStatus,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnExitSignal(signal string, coreDumped bool, errorMessage string, languageTag string) {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeExitSignal,
		Payload: message.PayloadExitSignal{
			Signal:       signal,
			CoreDumped:   coreDumped,
			ErrorMessage: errorMessage,
			LanguageTag:  languageTag,
		},
		ChannelID: l.channelID,
	})
}

func (l *loggerChannel) OnWriteClose() {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeWriteClose,
		ChannelID:    l.channelID,
	})
}

func (l *loggerChannel) OnClose() {
	l.c.log(message.Message{
		ConnectionID: l.c.connectionID,
		Timestamp:    time.Now().UnixNano(),
		MessageType:  message.TypeClose,
		ChannelID:    l.channelID,
	})
}

//endregion
