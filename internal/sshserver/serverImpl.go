package sshserver

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	ssh2 "github.com/containerssh/libcontainerssh/internal/ssh"
	"github.com/containerssh/libcontainerssh/log"
	messageCodes "github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/service"
	"golang.org/x/crypto/ssh"
)

type serverImpl struct {
	cfg                 config.SSHConfig
	logger              log.Logger
	handler             Handler
	listenSocket        net.Listener
	wg                  *sync.WaitGroup
	lock                *sync.Mutex
	clientSockets       map[*ssh.ServerConn]bool
	nextGlobalRequestID uint64
	nextChannelID       uint64
	hostKeys            []ssh.Signer
	shutdownHandlers    *shutdownRegistry
	shuttingDown        bool
}

func (s *serverImpl) String() string {
	return "SSH server"
}

func (s *serverImpl) RunWithLifecycle(lifecycle service.Lifecycle) error {
	s.lock.Lock()
	alreadyRunning := false
	if s.listenSocket != nil {
		alreadyRunning = true
	} else {
		s.clientSockets = make(map[*ssh.ServerConn]bool)
	}
	s.shuttingDown = false
	if alreadyRunning {
		s.lock.Unlock()
		return messageCodes.NewMessage(messageCodes.ESSHAlreadyRunning, "SSH server is already running")
	}

	listenConfig := net.ListenConfig{
		Control: s.socketControl,
	}

	netListener, err := listenConfig.Listen(lifecycle.Context(), "tcp", s.cfg.Listen)
	if err != nil {
		s.lock.Unlock()
		return messageCodes.Wrap(err, messageCodes.ESSHStartFailed, "failed to start SSH server on %s", s.cfg.Listen)
	}
	s.listenSocket = netListener
	s.lock.Unlock()
	if err := s.handler.OnReady(); err != nil {
		if err := netListener.Close(); err != nil {
			s.logger.Warning(
				messageCodes.Wrap(
					err,
					messageCodes.ESSHListenCloseFailed,
					"failed to close listen socket after failed startup",
				),
			)
		}
		return err
	}
	lifecycle.Running()
	s.logger.Info(messageCodes.NewMessage(messageCodes.MSSHServiceAvailable, "SSH server running on %s", s.cfg.Listen))

	go s.handleListenSocketOnShutdown(lifecycle)
	for {
		tcpConn, err := netListener.Accept()
		if err != nil {
			// Assume listen socket closed
			break
		}
		s.wg.Add(1)
		go s.handleConnection(tcpConn)
	}
	lifecycle.Stopping()
	s.shuttingDown = true
	allClientsExited := make(chan struct{})
	shutdownHandlerExited := make(chan struct{}, 1)
	go s.shutdownHandlers.Shutdown(lifecycle.ShutdownContext())
	go s.disconnectClients(lifecycle, allClientsExited)
	go s.shutdownHandler(lifecycle, shutdownHandlerExited)

	s.wg.Wait()
	close(allClientsExited)
	<-shutdownHandlerExited
	// nilerr will complain about this return becaus err may be not nil above, but that's not a problem since the
	// err only indicates an Accept failure.
	return nil //nolint:nilerr
}

func (s *serverImpl) handleListenSocketOnShutdown(lifecycle service.Lifecycle) {
	<-lifecycle.Context().Done()
	s.lock.Lock()
	if err := s.listenSocket.Close(); err != nil {
		s.logger.Warning(messageCodes.Wrap(err, messageCodes.ESSHListenCloseFailed, "failed to close listen socket"))
	}
	s.listenSocket = nil
	s.lock.Unlock()
}

func (s *serverImpl) disconnectClients(lifecycle service.Lifecycle, allClientsExited chan struct{}) {
	select {
	case <-allClientsExited:
		return
	case <-lifecycle.ShutdownContext().Done():
	}

	s.lock.Lock()
	for serverSocket := range s.clientSockets {
		_ = serverSocket.Close()
	}
	s.clientSockets = map[*ssh.ServerConn]bool{}
	s.lock.Unlock()
}

func (s *serverImpl) createPasswordAuthenticator(
	handlerNetworkConnection NetworkConnectionHandler,
	logger log.Logger,
) func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
	return func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
		authResponse, metadata, err := handlerNetworkConnection.OnAuthPassword(
			conn.User(),
			password,
			string(conn.ClientVersion()),
		)
		//goland:noinspection GoNilness
		switch authResponse {
		case AuthResponseSuccess:
			s.logAuthSuccessful(logger, conn, "Password")
			return &ssh.Permissions{}, metadata, nil
		case AuthResponseFailure:
			err = s.wrapAndLogAuthFailure(logger, conn, "Password", err)
			return nil, nil, err
		case AuthResponseUnavailable:
			err = s.wrapAndLogAuthUnavailable(logger, conn, "Password", err)
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("authentication currently unavailable")
	}
}

func (s *serverImpl) wrapAndLogAuthUnavailable(
	logger log.Logger,
	conn ssh.ConnMetadata,
	authMethod string,
	err error,
) error {
	err = messageCodes.WrapUser(
		err,
		messageCodes.ESSHAuthUnavailable,
		"Authentication is currently unavailable, please try a different authentication method or try again later.",
		"%s authentication for user %s currently unavailable.",
		authMethod,
		conn.User(),
	).
		Label("username", conn.User()).
		Label("method", strings.ToLower(authMethod)).
		Label("reason", err.Error())
	logger.Info(err)
	return err
}

func (s *serverImpl) wrapAndLogAuthFailure(
	logger log.Logger,
	conn ssh.ConnMetadata,
	authMethod string,
	err error,
) error {
	if err == nil {
		err = messageCodes.UserMessage(
			messageCodes.ESSHAuthFailed,
			"Authentication failed.",
			"%s authentication for user %s failed.",
			authMethod,
			conn.User(),
		).
			Label("username", conn.User()).
			Label("method", strings.ToLower(authMethod))
		logger.Info(err)
	} else {
		err = messageCodes.WrapUser(
			err,
			messageCodes.ESSHAuthFailed,
			"Authentication failed.",
			"%s authentication for user %s failed.",
			authMethod,
			conn.User(),
		).
			Label("username", conn.User()).
			Label("method", strings.ToLower(authMethod)).
			Label("reason", err.Error())
		logger.Info(err)
	}
	return err
}

func (s *serverImpl) logAuthSuccessful(logger log.Logger, conn ssh.ConnMetadata, authMethod string) {
	err := messageCodes.UserMessage(
		messageCodes.ESSHAuthSuccessful,
		"Authentication successful.",
		"%s authentication for user %s successful.",
		authMethod,
		conn.User(),
	).Label("username", conn.User()).Label("method", strings.ToLower(authMethod))
	logger.Info(err)
}

func (s *serverImpl) createPubKeyAuthenticator(
	handlerNetworkConnection NetworkConnectionHandler,
	logger log.Logger,
) func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
	return func(conn ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
		authorizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubKey)))
		authResponse, metadata, err := handlerNetworkConnection.OnAuthPubKey(
			conn.User(),
			authorizedKey,
			string(conn.ClientVersion()),
		)
		//goland:noinspection GoNilness
		switch authResponse {
		case AuthResponseSuccess:
			s.logAuthSuccessful(logger, conn, "Public key")
			return &ssh.Permissions{}, metadata, nil
		case AuthResponseFailure:
			err = s.wrapAndLogAuthFailure(logger, conn, "Public key", err)
			return nil, nil, err
		case AuthResponseUnavailable:
			err = s.wrapAndLogAuthUnavailable(logger, conn, "Public key", err)
			return nil, nil, err
		}
		// This should never happen
		return nil, nil, fmt.Errorf("authentication currently unavailable")
	}
}

func (s *serverImpl) createKeyboardInteractiveHandler(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
	return func(
		conn ssh.ConnMetadata,
		challenge ssh.KeyboardInteractiveChallenge,
	) (*ssh.Permissions, *auth.ConnectionMetadata, error) {
		challengeWrapper := func(
			instruction string,
			questions KeyboardInteractiveQuestions,
		) (answers KeyboardInteractiveAnswers, err error) {
			if answers.answers == nil {
				answers.answers = map[string]string{}
			}
			var q []string
			var echos []bool
			for _, question := range questions {
				q = append(q, question.Question)
				echos = append(echos, question.EchoResponse)
			}

			// user, instruction string, questions []string, echos []bool
			answerList, err := challenge(conn.User(), instruction, q, echos)
			for index, rawAnswer := range answerList {
				question := questions[index]
				answers.answers[question.getID()] = rawAnswer
			}
			return answers, err
		}
		authResponse, metadata, err := handlerNetworkConnection.OnAuthKeyboardInteractive(
			conn.User(),
			challengeWrapper,
			string(conn.ClientVersion()),
		)
		//goland:noinspection GoNilness
		switch authResponse {
		case AuthResponseSuccess:
			s.logAuthSuccessful(logger, conn, "Keyboard-interactive")
			return &ssh.Permissions{}, metadata, nil
		case AuthResponseFailure:
			err = s.wrapAndLogAuthFailure(logger, conn, "Keyboard-interactive", err)
			return nil, nil, err
		case AuthResponseUnavailable:
			err = s.wrapAndLogAuthUnavailable(logger, conn, "Keyboard-interactive", err)
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("authentication currently unavailable")
	}
}

func (s *serverImpl) createConfiguration(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) *ssh.ServerConfig {
	passwordCallback, pubkeyCallback, keyboardInteractiveCallback, gssConfig := s.createAuthenticators(
		handlerNetworkConnection,
		logger,
	)

	serverConfig := &ssh.ServerConfig{
		Config: ssh.Config{
			KeyExchanges: s.cfg.KexAlgorithms.StringList(),
			Ciphers:      s.cfg.Ciphers.StringList(),
			MACs:         s.cfg.MACs.StringList(),
		},
		NoClientAuth:                false,
		MaxAuthTries:                6,
		PasswordCallback:            passwordCallback,
		PublicKeyCallback:           pubkeyCallback,
		KeyboardInteractiveCallback: keyboardInteractiveCallback,
		GSSAPIWithMICConfig: gssConfig,
		ServerVersion:               s.cfg.ServerVersion.String(),
		BannerCallback:              func(conn ssh.ConnMetadata) string { return s.cfg.Banner },
	}
	for _, key := range s.hostKeys {
		serverConfig.AddHostKey(key)
	}
	return serverConfig
}

func (s *serverImpl) createAuthenticators(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) (
	func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error),
	func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error),
	func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error),
	*ssh.GSSAPIWithMICConfig,
) {
	passwordCallback := s.createPasswordCallback(handlerNetworkConnection, logger)
	pubkeyCallback := s.createPubKeyCallback(handlerNetworkConnection, logger)
	keyboardInteractiveCallback := s.createKeyboardInteractiveCallback(handlerNetworkConnection, logger)
	gssConfig := s.createGSSAPIConfig(handlerNetworkConnection, logger)
	return passwordCallback, pubkeyCallback, keyboardInteractiveCallback, gssConfig
}

func (s *serverImpl) createGSSAPIConfig(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) (*ssh.GSSAPIWithMICConfig){
	var gssConfig *ssh.GSSAPIWithMICConfig

	gssServer := handlerNetworkConnection.OnAuthGSSAPI()
	if gssServer != nil {
		gssConfig = &ssh.GSSAPIWithMICConfig{
			AllowLogin: func(conn ssh.ConnMetadata, srcName string) (*ssh.Permissions, error) {
				if !gssServer.Success() {
					if gssServer.Error() == nil {
						return nil, messageCodes.NewMessage(
							messageCodes.ESSHAuthFailed,
							"Authentication failed",
						)
					}
					return nil, gssServer.Error()
				}

				if err := gssServer.AllowLogin(conn.User()); err != nil {
					return nil, err
				}

				metadata := gssServer.Metadata()

				s.logAuthSuccessful(logger, conn, "GSSAPI")
				sshConnectionHandler, err := handlerNetworkConnection.OnHandshakeSuccess(
					conn.User(),
					string(conn.ClientVersion()),
					metadata,
				)
				if err != nil {
					err = messageCodes.WrapUser(
						err,
						messageCodes.ESSHBackendRejected,
						"Authentication currently unavailable, please try again later.",
						"The backend has rejected the user after successful authentication.",
					)
					logger.Error(err)
					return nil, err
				}
				handlerNetworkConnection.sshConnectionHandler = sshConnectionHandler
				return &ssh.Permissions{}, nil
			},
			Server: gssServer,
		}
	} else {
		logger.Info(
			messageCodes.NewMessage(
				messageCodes.ESSHAuthUnavailable,
				"GSSAPI Authentication unsupported with current authentication method",
			),
		)
	}
	return gssConfig
}

func (s *serverImpl) createKeyboardInteractiveCallback(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	keyboardInteractiveHandler := s.createKeyboardInteractiveHandler(handlerNetworkConnection, logger)
	keyboardInteractiveCallback := func(
		conn ssh.ConnMetadata,
		challenge ssh.KeyboardInteractiveChallenge,
	) (*ssh.Permissions, error) {
		permissions, metadata, err := keyboardInteractiveHandler(conn, challenge)
		if err != nil {
			return permissions, err
		}
		// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
		sshConnectionHandler, err := handlerNetworkConnection.OnHandshakeSuccess(
			conn.User(),
			string(conn.ClientVersion()),
			metadata,
		)
		if err != nil {
			err = messageCodes.WrapUser(
				err,
				messageCodes.ESSHBackendRejected,
				"Authentication currently unavailable, please try again later.",
				"The backend has rejected the user after successful authentication.",
			)
			logger.Error(err)
			return permissions, err
		}
		handlerNetworkConnection.sshConnectionHandler = sshConnectionHandler
		return permissions, err
	}
	return keyboardInteractiveCallback
}

func (s *serverImpl) createPubKeyCallback(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	pubKeyHandler := s.createPubKeyAuthenticator(handlerNetworkConnection, logger)
	pubkeyCallback := func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		permissions, metadata, err := pubKeyHandler(conn, key)
		if err != nil {
			return permissions, err
		}
		// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
		sshConnectionHandler, err := handlerNetworkConnection.OnHandshakeSuccess(
			conn.User(),
			string(conn.ClientVersion()),
			metadata,
		)
		if err != nil {
			err = messageCodes.WrapUser(
				err,
				messageCodes.ESSHBackendRejected,
				"Authentication currently unavailable, please try again later.",
				"The backend has rejected the user after successful authentication.",
			)
			logger.Error(err)
			return permissions, err
		}
		handlerNetworkConnection.sshConnectionHandler = sshConnectionHandler
		return permissions, err
	}
	return pubkeyCallback
}

func (s *serverImpl) createPasswordCallback(
	handlerNetworkConnection *networkConnectionWrapper,
	logger log.Logger,
) func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	passwordHandler := s.createPasswordAuthenticator(handlerNetworkConnection, logger)
	passwordCallback := func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		permissions, metadata, err := passwordHandler(conn, password)
		if err != nil {
			return permissions, err
		}
		// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
		sshConnectionHandler, err := handlerNetworkConnection.OnHandshakeSuccess(
			conn.User(),
			string(conn.ClientVersion()),
			metadata,
		)
		if err != nil {
			err = messageCodes.WrapUser(
				err,
				messageCodes.ESSHBackendRejected,
				"Authentication currently unavailable, please try again later.",
				"The backend has rejected the user after successful authentication.",
			)
			logger.Error(err)
			return permissions, err
		}
		handlerNetworkConnection.sshConnectionHandler = sshConnectionHandler
		return permissions, err
	}
	return passwordCallback
}

func (s *serverImpl) handleConnection(conn net.Conn) {
	addr := conn.RemoteAddr().(*net.TCPAddr)
	connectionID := GenerateConnectionID()
	logger := s.logger.
		WithLabel("remoteAddr", addr.IP.String()).
		WithLabel("connectionId", connectionID)
	handlerNetworkConnection, err := s.handler.OnNetworkConnection(*addr, connectionID)
	if err != nil {
		logger.Info(err)
		_ = conn.Close()
		s.wg.Done()
		return
	}
	shutdownHandlerID := fmt.Sprintf("network-%s", connectionID)
	s.shutdownHandlers.Register(shutdownHandlerID, handlerNetworkConnection)

	logger.Debug(
		messageCodes.NewMessage(
			messageCodes.MSSHConnected, "Client connected",
		),
	)

	// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
	wrapper := networkConnectionWrapper{
		NetworkConnectionHandler: handlerNetworkConnection,
	}

	sshConn, channels, globalRequests, err := ssh.NewServerConn(conn, s.createConfiguration(&wrapper, logger))
	if err != nil {
		logger.Info(messageCodes.Wrap(err, messageCodes.ESSHHandshakeFailed, "SSH handshake failed"))
		handlerNetworkConnection.OnHandshakeFailed(err)
		s.shutdownHandlers.Unregister(shutdownHandlerID)
		logger.Debug(messageCodes.NewMessage(messageCodes.MSSHDisconnected, "Client disconnected"))
		handlerNetworkConnection.OnDisconnect()
		_ = conn.Close()
		s.wg.Done()
		return
	}
	logger = logger.WithLabel("username", sshConn.User())
	logger.Debug(messageCodes.NewMessage(messageCodes.MSSHHandshakeSuccessful, "SSH handshake successful"))
	s.lock.Lock()
	s.clientSockets[sshConn] = true
	sshShutdownHandlerID := fmt.Sprintf("ssh-%s", connectionID)
	s.lock.Unlock()

	if s.cfg.ClientAliveInterval > 0 {
		go func() {
			missedAlives := 0
			for {
				time.Sleep(s.cfg.ClientAliveInterval)

				_, _, err := sshConn.SendRequest("keepalive@openssh.com", true, []byte{})

				if err != nil {
					missedAlives++

					logger.Debug(
						messageCodes.Wrap(
							err,
							messageCodes.ESSHKeepAliveFailed,
							"Keepalive error",
						),
					)
					if missedAlives >= s.cfg.ClientAliveCountMax {
						_ = sshConn.Close()
						break
					}
					continue
				}
				missedAlives = 0
			}
		}()
	}

	go func() {
		_ = sshConn.Wait()
		logger.Debug(messageCodes.NewMessage(messageCodes.MSSHDisconnected, "Client disconnected"))
		s.shutdownHandlers.Unregister(shutdownHandlerID)
		s.shutdownHandlers.Unregister(sshShutdownHandlerID)
		handlerNetworkConnection.OnDisconnect()
		s.wg.Done()
	}()
	// HACK: check HACKS.md "OnHandshakeSuccess conformanceTestHandler"
	handlerSSHConnection := wrapper.sshConnectionHandler
	s.shutdownHandlers.Register(sshShutdownHandlerID, handlerSSHConnection)

	go s.handleChannels(connectionID, channels, handlerSSHConnection, logger)
	go s.handleGlobalRequests(globalRequests, handlerSSHConnection, logger)
}

func (s *serverImpl) handleKeepAliveRequest(req *ssh.Request, logger log.Logger) {
	if req.WantReply {
		if err := req.Reply(false, []byte{}); err != nil {
			logger.Debug(
				messageCodes.Wrap(
					err,
					messageCodes.ESSHReplyFailed,
					"failed to send reply to global request type %s",
					req.Type,
				),
			)
		}
	}
}

func (s *serverImpl) handleUnknownGlobalRequest(req *ssh.Request, requestID uint64, connection SSHConnectionHandler, logger log.Logger) {
	logger.Debug(
		messageCodes.NewMessage(messageCodes.ESSHUnsupportedGlobalRequest, "Unsupported global request").Label(
			"type",
			req.Type,
		),
	)

	connection.OnUnsupportedGlobalRequest(requestID, req.Type, req.Payload)
	if req.WantReply {
		if err := req.Reply(false, []byte("request type not supported")); err != nil {
			logger.Debug(
				messageCodes.Wrap(
					err,
					messageCodes.ESSHReplyFailed,
					"failed to send reply to global request type %s",
					req.Type,
				),
			)
		}
	}
}

func (s *serverImpl) handleGlobalRequests(
	requests <-chan *ssh.Request,
	connection SSHConnectionHandler,
	logger log.Logger,
) {
	for {
		request, ok := <-requests
		if !ok {
			break
		}
		requestID := s.nextGlobalRequestID
		s.nextGlobalRequestID++

		switch request.Type {
		case "keepalive@openssh.com":
			s.handleKeepAliveRequest(request, logger)
		default:
			s.handleUnknownGlobalRequest(request, requestID, connection, logger)
		}
	}
}

func (s *serverImpl) handleChannels(
	connectionID string,
	channels <-chan ssh.NewChannel,
	connection SSHConnectionHandler,
	logger log.Logger,
) {
	for {
		newChannel, ok := <-channels
		if !ok {
			break
		}
		s.lock.Lock()
		channelID := s.nextChannelID
		s.nextChannelID++
		s.lock.Unlock()
		logger = logger.WithLabel("channelId", channelID)
		if newChannel.ChannelType() != "session" {
			logger.Debug(
				messageCodes.NewMessage(
					messageCodes.ESSHUnsupportedChannelType,
					"Unsupported channel type requested",
				).Label("type", newChannel.ChannelType()),
			)
			connection.OnUnsupportedChannel(channelID, newChannel.ChannelType(), newChannel.ExtraData())
			if err := newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type"); err != nil {
				logger.Debug("failed to send channel rejection for channel type %s", newChannel.ChannelType())
			}
			continue
		}
		go s.handleSessionChannel(connectionID, channelID, newChannel, connection, logger)
	}
}

func (s *serverImpl) handleSessionChannel(
	connectionID string,
	channelID uint64,
	newChannel ssh.NewChannel,
	connection SSHConnectionHandler,
	logger log.Logger,
) {
	channelCallbacks := &channelWrapper{
		logger: logger,
		lock:   &sync.Mutex{},
	}
	handlerChannel, rejection := connection.OnSessionChannel(channelID, newChannel.ExtraData(), channelCallbacks)
	if rejection != nil {
		logger.Debug(
			messageCodes.Wrap(
				rejection,
				messageCodes.MSSHNewChannelRejected,
				"New SSH channel rejected",
			).Label("type", newChannel.ChannelType()),
		)

		if err := newChannel.Reject(rejection.Reason(), rejection.UserMessage()); err != nil {
			logger.Debug(messageCodes.Wrap(err, messageCodes.ESSHReplyFailed, "Failed to send reply to channel request"))
		}
		return
	}
	shutdownHandlerID := fmt.Sprintf("session-%s-%d", connectionID, channelID)
	s.shutdownHandlers.Register(shutdownHandlerID, handlerChannel)
	channel, requests, err := newChannel.Accept()
	if err != nil {
		logger.Debug(messageCodes.Wrap(err, messageCodes.ESSHReplyFailed, "failed to accept session channel"))
		s.shutdownHandlers.Unregister(shutdownHandlerID)
		return
	}
	logger.Debug(messageCodes.NewMessage(messageCodes.MSSHNewChannel, "New SSH channel").Label("type", newChannel.ChannelType()))
	channelCallbacks.channel = channel
	nextRequestID := uint64(0)
	for {
		request, ok := <-requests
		if !ok {
			s.shutdownHandlers.Unregister(shutdownHandlerID)
			channelCallbacks.onClose()
			handlerChannel.OnClose()
			break
		}
		requestID := nextRequestID
		nextRequestID++
		s.handleChannelRequest(requestID, request, handlerChannel, logger)
	}
}

func (s *serverImpl) unmarshalEnv(request *ssh.Request) (payload ssh2.EnvRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalPty(request *ssh.Request) (payload ssh2.PtyRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalShell(request *ssh.Request) (payload ssh2.ShellRequestPayload, err error) {
	if len(request.Payload) != 0 {
		err = ssh.Unmarshal(request.Payload, &payload)
	}
	return payload, err
}

func (s *serverImpl) unmarshalExec(request *ssh.Request) (payload ssh2.ExecRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalSubsystem(request *ssh.Request) (payload ssh2.SubsystemRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalWindow(request *ssh.Request) (payload ssh2.WindowRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalSignal(request *ssh.Request) (payload ssh2.SignalRequestPayload, err error) {
	return payload, ssh.Unmarshal(request.Payload, &payload)
}

func (s *serverImpl) unmarshalPayload(request *ssh.Request) (payload interface{}, err error) {
	switch ssh2.RequestType(request.Type) {
	case ssh2.RequestTypeEnv:
		return s.unmarshalEnv(request)
	case ssh2.RequestTypePty:
		return s.unmarshalPty(request)
	case ssh2.RequestTypeShell:
		return s.unmarshalShell(request)
	case ssh2.RequestTypeExec:
		return s.unmarshalExec(request)
	case ssh2.RequestTypeSubsystem:
		return s.unmarshalSubsystem(request)
	case ssh2.RequestTypeWindow:
		return s.unmarshalWindow(request)
	case ssh2.RequestTypeSignal:
		return s.unmarshalSignal(request)
	default:
		return nil, nil
	}
}

func (s *serverImpl) handleChannelRequest(
	requestID uint64,
	request *ssh.Request,
	sessionChannel SessionChannelHandler,
	logger log.Logger,
) {
	reply := s.createReply(request, logger)
	payload, err := s.unmarshalPayload(request)
	if payload == nil {
		sessionChannel.OnUnsupportedChannelRequest(requestID, request.Type, request.Payload)
		reply(false, fmt.Sprintf("unsupported request type: %s", request.Type), nil)
		return
	}
	if err != nil {
		logger.Debug(
			messageCodes.Wrap(
				err,
				messageCodes.ESSHDecodeFailed,
				"failed to unmarshal %s request payload",
				request.Type,
			),
		)
		sessionChannel.OnFailedDecodeChannelRequest(requestID, request.Type, request.Payload, err)
		reply(false, "failed to unmarshal payload", nil)
		return
	}
	logger.Debug(
		messageCodes.NewMessage(
			messageCodes.MSSHChannelRequest,
			"%s channel request from client",
			request.Type,
		).Label("RequestType", request.Type),
	)
	if err := s.handleDecodedChannelRequest(
		requestID,
		ssh2.RequestType(request.Type),
		payload,
		sessionChannel,
	); err != nil {
		logger.Debug(
			messageCodes.NewMessage(
				messageCodes.MSSHChannelRequestFailed,
				"%s channel request from client failed",
				request.Type,
			).Label("RequestType", request.Type),
		)
		reply(false, err.Error(), err)
		return
	}
	logger.Debug(
		messageCodes.NewMessage(
			messageCodes.MSSHChannelRequestSuccessful,
			"%s channel request from client successful",
			request.Type,
		).Label("RequestType", request.Type),
	)
	reply(true, "", nil)
}

func (s *serverImpl) createReply(request *ssh.Request, logger log.Logger) func(
	success bool,
	message string,
	reason error,
) {
	reply := func(success bool, message string, reason error) {
		if request.WantReply {
			if err := request.Reply(
				success,
				[]byte(message),
			); err != nil {
				logger.Debug(
					messageCodes.Wrap(
						err,
						messageCodes.ESSHReplyFailed,
						"Failed to send reply to client",
					),
				)
			}
		}
	}
	return reply
}

func (s *serverImpl) handleDecodedChannelRequest(
	requestID uint64,
	requestType ssh2.RequestType,
	payload interface{},
	sessionChannel SessionChannelHandler,
) error {
	switch requestType {
	case ssh2.RequestTypeEnv:
		return s.onEnvRequest(requestID, sessionChannel, payload)
	case ssh2.RequestTypePty:
		return s.onPtyRequest(requestID, sessionChannel, payload)
	case ssh2.RequestTypeShell:
		return s.onShell(requestID, sessionChannel)
	case ssh2.RequestTypeExec:
		return s.onExec(requestID, sessionChannel, payload)
	case ssh2.RequestTypeSubsystem:
		return s.onSubsystem(requestID, sessionChannel, payload)
	case ssh2.RequestTypeWindow:
		return s.onChannel(requestID, sessionChannel, payload)
	case ssh2.RequestTypeSignal:
		return s.onSignal(requestID, sessionChannel, payload)
	}
	return nil
}

func (s *serverImpl) onEnvRequest(requestID uint64, sessionChannel SessionChannelHandler, payload interface{}) error {
	return sessionChannel.OnEnvRequest(
		requestID,
		payload.(ssh2.EnvRequestPayload).Name,
		payload.(ssh2.EnvRequestPayload).Value,
	)
}

func (s *serverImpl) onPtyRequest(requestID uint64, sessionChannel SessionChannelHandler, payload interface{}) error {
	return sessionChannel.OnPtyRequest(
		requestID,
		payload.(ssh2.PtyRequestPayload).Term,
		payload.(ssh2.PtyRequestPayload).Columns,
		payload.(ssh2.PtyRequestPayload).Rows,
		payload.(ssh2.PtyRequestPayload).Width,
		payload.(ssh2.PtyRequestPayload).Height,
		payload.(ssh2.PtyRequestPayload).ModeList,
	)
}

func (s *serverImpl) onShell(
	requestID uint64,
	sessionChannel SessionChannelHandler,
) error {
	return sessionChannel.OnShell(
		requestID,
	)
}

func (s *serverImpl) onExec(
	requestID uint64,
	sessionChannel SessionChannelHandler,
	payload interface{},
) error {
	return sessionChannel.OnExecRequest(
		requestID,
		payload.(ssh2.ExecRequestPayload).Exec,
	)
}

func (s *serverImpl) onSignal(requestID uint64, sessionChannel SessionChannelHandler, payload interface{}) error {
	return sessionChannel.OnSignal(
		requestID,
		payload.(ssh2.SignalRequestPayload).Signal,
	)
}

func (s *serverImpl) onSubsystem(
	requestID uint64,
	sessionChannel SessionChannelHandler,
	payload interface{},
) error {
	return sessionChannel.OnSubsystem(
		requestID,
		payload.(ssh2.SubsystemRequestPayload).Subsystem,
	)
}

func (s *serverImpl) onChannel(requestID uint64, sessionChannel SessionChannelHandler, payload interface{}) error {
	return sessionChannel.OnWindow(
		requestID,
		payload.(ssh2.WindowRequestPayload).Columns,
		payload.(ssh2.WindowRequestPayload).Rows,
		payload.(ssh2.WindowRequestPayload).Width,
		payload.(ssh2.WindowRequestPayload).Height,
	)
}

func (s *serverImpl) shutdownHandler(lifecycle service.Lifecycle, exited chan struct{}) {
	s.handler.OnShutdown(lifecycle.ShutdownContext())
	exited <- struct{}{}
}

type shutdownHandler interface {
	OnShutdown(shutdownContext context.Context)
}
