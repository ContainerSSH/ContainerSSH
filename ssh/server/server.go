package server

import (
	"context"
	"fmt"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/protocol"
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/metrics"
	"golang.org/x/crypto/ssh"
	"net"
)

var ErrorAuthenticationFailed = fmt.Errorf("authentication failed")

type RequestResponse struct {
	Success bool
	Payload []byte
}

type ReadyHandler interface {
	// This func will be called once the SSH server is ready to receive connections.
	OnReady(listener net.Listener)
}

type ConnectionHandler interface {
	OnConnection(*ssh.ServerConn, *audit.Connection) (GlobalRequestHandler, ChannelHandler, error)
}

type GlobalRequestHandler interface {
	OnGlobalRequest(
		ctx context.Context,
		sshConn *ssh.ServerConn,
		requestType string,
		payload []byte,
	) RequestResponse
}

type ChannelRejection struct {
	RejectionReason  ssh.RejectionReason
	RejectionMessage string
}

type ChannelHandler interface {
	// This func will be called when a new session channel is requested and gives an opportunity to decide
	// if the channel should be opened. If an error is returned the channel is rejected, otherwise it is accepted.
	// This function should NOT handle the channel itself.
	OnChannel(
		ctx context.Context,
		connection ssh.ConnMetadata,
		channelType string,
		extraData []byte,
	) (ChannelRequestHandler, *ChannelRejection)
}

type ChannelRequestHandler interface {
	// This func will be called when a new requests arrives in a channel. This func can handle the request and return
	// a response.
	OnChannelRequest(ctx context.Context, sshConn *ssh.ServerConn, channel ssh.Channel, requestType string, payload []byte) RequestResponse
}

type Config struct {
	ssh.Config

	HostKeys     []ssh.Signer
	NoClientAuth bool

	MaxAuthTries      int
	PasswordCallback  func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)
	PublicKeyCallback func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error)
	AuthLogCallback   func(conn ssh.ConnMetadata, method string, err error)
	ServerVersion     string
	BannerCallback    func(conn ssh.ConnMetadata) string
}

type Server struct {
	listen            string
	serverConfig      *Config
	readyHandler      ReadyHandler
	connectionHandler ConnectionHandler
	logger            log.Logger
	metric            *metrics.MetricCollector
	audit             audit.Plugin
	auditConfig       config.AuditConfig
}

func New(
	listen string,
	serverConfig *Config,
	readyHandler ReadyHandler,
	connectionHandler ConnectionHandler,
	logger log.Logger,
	metric *metrics.MetricCollector,
	audit audit.Plugin,
	auditConfig config.AuditConfig,
) (*Server, error) {
	server := &Server{
		listen:            listen,
		serverConfig:      serverConfig,
		readyHandler:      readyHandler,
		connectionHandler: connectionHandler,
		logger:            logger,
		metric:            metric,
		audit:             audit,
		auditConfig:       auditConfig,
	}

	err := server.validateConfig()
	if err != nil {
		return nil, err
	}

	metric.SetMetricMeta(MetricNameConnections, "Number of connections since start", metrics.MetricTypeCounter)
	metric.SetMetricMeta(MetricNameSuccessfulHandshake, "Successful SSH handshakes since start", metrics.MetricTypeCounter)
	metric.SetMetricMeta(MetricNameFailedHandshake, "Failed SSH handshakes since start", metrics.MetricTypeCounter)
	metric.SetMetricMeta(MetricNameCurrentConnections, "Current open SSH connections", metrics.MetricTypeGauge)
	return server, err
}

var supportedCiphers = []string{
	"aes128-ctr", "aes192-ctr", "aes256-ctr",
	"aes128-gcm@openssh.com",
	"chacha20-poly1305@openssh.com",
	"arcfour256", "arcfour128", "arcfour",
	"aes128-cbc",
	"tripledescbcID",
}
var supportedKexAlgos = []string{
	"curve25519-sha256@libssh.org",
	"ecdh-sha2-nistp256", "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
	"diffie-hellman-group14-sha1", "diffie-hellman-group1-sha1",
}
var supportedHostKeyAlgos = []string{
	"ssh-rsa-cert-v01@openssh.com", "ssh-dss-cert-v01@openssh.com", "ecdsa-sha2-nistp256-cert-v01@openssh.com",
	"ecdsa-sha2-nistp384-cert-v01@openssh.com", "ecdsa-sha2-nistp521-cert-v01@openssh.com",
	"ssh-ed25519-cert-v01@openssh.com",

	"ecdsa-sha2-nistp256-cert-v01@openssh.com", "ecdsa-sha2-nistp384-cert-v01@openssh.com",
	"ecdsa-sha2-nistp521-cert-v01@openssh.com",
	"ssh-rsa", "ssh-dss",

	"ssh-ed25519",
}
var supportedMACs = []string{
	"hmac-sha2-256-etm@openssh.com", "hmac-sha2-256", "hmac-sha1", "hmac-sha1-96",
}

func (server *Server) findUnsupported(name string, requestedList []string, supportedList []string) error {
	for _, requestedItem := range requestedList {
		found := false
		for _, supportedItem := range supportedList {
			if supportedItem == requestedItem {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ssh: unsupported %s %s for server", name, requestedItem)
		}
	}
	return nil
}

func (server *Server) validateCiphers(config *Config) error {
	if len(config.Ciphers) == 0 {
		return nil
	}
	return server.findUnsupported("cipher", config.Ciphers, supportedCiphers)
}

func (server *Server) validateKexAlgorithms(config *Config) error {
	if len(config.KeyExchanges) == 0 {
		return nil
	}
	return server.findUnsupported("key exchange algorithm", config.KeyExchanges, supportedKexAlgos)
}

func (server *Server) validateMACs(config *Config) error {
	if len(config.MACs) == 0 {
		return nil
	}
	return server.findUnsupported("MAC algorithm", config.MACs, supportedMACs)
}

func (server *Server) validateHostKeys(config *Config) error {
	if len(config.HostKeys) == 0 {
		return fmt.Errorf("no host keys supplied")
	}
	for index, hostKey := range config.HostKeys {
		if hostKey == nil {
			return fmt.Errorf("host key %d is nil (probably not loaded correctly)", index)
		}
		foundHostKeyAlgo := false
		for _, hostKeyAlgo := range supportedHostKeyAlgos {
			if hostKey.PublicKey().Type() == hostKeyAlgo {
				foundHostKeyAlgo = true
			}
		}
		if !foundHostKeyAlgo {
			return fmt.Errorf("unknown host key format (%s)", hostKey.PublicKey().Type())
		}
	}
	return nil
}

func (server *Server) validateConfig() error {
	validators := []func(config *Config) error{
		server.validateHostKeys,
		server.validateCiphers,
		server.validateKexAlgorithms,
		server.validateMACs,
	}

	for _, validator := range validators {
		err := validator(server.serverConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server *Server) createConfig(auditConnection *audit.Connection) *ssh.ServerConfig {
	cfg := &ssh.ServerConfig{
		Config:       server.serverConfig.Config,
		NoClientAuth: server.serverConfig.NoClientAuth,
		MaxAuthTries: server.serverConfig.MaxAuthTries,
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			var auditPassword []byte
			if auditConnection.Intercept.Passwords {
				auditPassword = password
			}
			auditConnection.Message(protocol.MessageType_AuthPassword, protocol.PayloadAuthPassword{
				Username: conn.User(),
				Password: auditPassword,
			})
			permissions, err := server.serverConfig.PasswordCallback(conn, password)
			if err != nil {
				if err == ErrorAuthenticationFailed {
					auditConnection.Message(protocol.MessageType_AuthPasswordFailed, protocol.PayloadAuthPassword{
						Username: conn.User(),
						Password: auditPassword,
					})
				} else {
					auditConnection.Message(protocol.MessageType_AuthPasswordBackendError, protocol.PayloadAuthPassword{
						Username: conn.User(),
						Password: auditPassword,
					})
				}
			} else {
				auditConnection.Message(protocol.MessageType_AuthPasswordSuccessful, protocol.PayloadAuthPassword{
					Username: conn.User(),
					Password: auditPassword,
				})
			}
			return permissions, err
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			auditConnection.Message(protocol.MessageType_AuthPubKey, protocol.PayloadAuthPubKey{
				Username: conn.User(),
				Key:      key.Marshal(),
			})
			permissions, err := server.serverConfig.PublicKeyCallback(conn, key)
			if err != nil {
				if err == ErrorAuthenticationFailed {
					auditConnection.Message(protocol.MessageType_AuthPubKeyFailed, protocol.PayloadAuthPubKey{
						Username: conn.User(),
						Key:      key.Marshal(),
					})
				} else {
					auditConnection.Message(protocol.MessageType_AuthPubKeyBackendError, protocol.PayloadAuthPubKey{
						Username: conn.User(),
						Key:      key.Marshal(),
					})
				}
			} else {
				auditConnection.Message(protocol.MessageType_AuthPubKeySuccessful, protocol.PayloadAuthPubKey{
					Username: conn.User(),
					Key:      key.Marshal(),
				})
			}
			return permissions, err
		},
		AuthLogCallback: server.serverConfig.AuthLogCallback,
		ServerVersion:   server.serverConfig.ServerVersion,
		BannerCallback:  server.serverConfig.BannerCallback,
	}

	for _, hostKey := range server.serverConfig.HostKeys {
		cfg.AddHostKey(hostKey)
	}

	return cfg
}

func (server *Server) Run(ctx context.Context) error {

	server.logger.InfoF("starting SSH server on %s", server.listen)
	netListener, err := net.Listen("tcp", server.listen)
	if err != nil {
		return err
	}
	if server.readyHandler != nil {
		server.readyHandler.OnReady(netListener)
	}
	go func() {
		for {
			tcpConn, err := netListener.Accept()
			if err != nil {
				// Assume listen socket closed
				break
			}
			ip := net.ParseIP(tcpConn.RemoteAddr().String())
			auditConnection, err := audit.GetConnection(server.audit, server.auditConfig)
			if err != nil {
				server.logger.ErrorF("failed to get random ID for connection (%v)", err)
				_ = tcpConn.Close()
				continue
			}
			auditConnection.Message(protocol.MessageType_Connect, protocol.PayloadConnect{
				RemoteAddr: ip.String(),
			})
			server.metric.IncrementGeo(MetricConnections, ip)
			server.logger.DebugF("connection from: %s", tcpConn.RemoteAddr().String())

			sshConfig := server.createConfig(auditConnection)
			sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, sshConfig)
			if err != nil {
				server.metric.IncrementGeo(MetricFailedHandshake, ip)
				server.logger.DebugF("failed handshake (%v)", err)
				auditConnection.Message(protocol.MessageType_Disconnect, nil)
				continue
			}
			server.logger.DebugF("new SSH connection from %s for user %s (%s)", sshConn.RemoteAddr(), sshConn.User(), sshConn.ClientVersion())
			server.metric.IncrementGeo(MetricSuccessfulHandshake, ip)
			server.metric.IncrementGeo(MetricCurrentConnections, ip)

			go func() {
				_ = sshConn.Wait()
				auditConnection.Message(protocol.MessageType_Disconnect, nil)
				server.metric.DecrementGeo(MetricCurrentConnections, ip)
			}()

			if server.connectionHandler == nil {
				server.logger.DebugF("no connection handler defined, closing connection")
				err = sshConn.Close()
				if err != nil {
					server.logger.DebugF("failed to close newly opened connection (%v)", err)
				}
				continue
			}

			globalRequestHandler, channelHandler, err := server.connectionHandler.OnConnection(sshConn, auditConnection)
			if err != nil {
				server.logger.DebugF("error from connection handler (%v)", err)
				err = sshConn.Close()
				if err != nil {
					server.logger.DebugF("failed to close newly opened connection (%v)", err)
				}
				continue
			}

			go server.handleGlobalRequests(ctx, globalRequestHandler, sshConn, reqs)
			go server.handleChannels(ctx, channelHandler, sshConn, chans, auditConnection)
		}
	}()

	<-ctx.Done()
	err = netListener.Close()
	if err != nil {
		server.logger.WarningF("failed to close listen socket (%v)", err)
		return err
	}

	return nil
}

func (server Server) handleGlobalRequests(ctx context.Context, globalRequestHandler GlobalRequestHandler, sshConn *ssh.ServerConn, reqs <-chan *ssh.Request) {
	for req := range reqs {
		globalRequest := req
		if globalRequestHandler == nil {
			server.replyRequest(globalRequest, false, []byte(fmt.Sprintf("unknown request type (%s)", req.Type)))
			continue
		}
		go func() {
			response := globalRequestHandler.OnGlobalRequest(ctx, sshConn, globalRequest.Type, globalRequest.Payload)
			if response.Success {
				server.replyRequest(globalRequest, true, response.Payload)
			} else {
				server.logger.DebugF("global request globalRequestHandler failed (%v)", response.Payload)
				server.replyRequest(globalRequest, false, response.Payload)
			}
		}()
	}
}

func (server *Server) handleChannels(
	ctx context.Context,
	channelHandler ChannelHandler,
	sshConn *ssh.ServerConn,
	chans <-chan ssh.NewChannel,
	auditConnection *audit.Connection,
) {
	for newChannel := range chans {
		go server.handleChannel(ctx, channelHandler, sshConn, newChannel, auditConnection)
	}
}

func (server *Server) handleChannel(
	ctx context.Context,
	channelHandler ChannelHandler,
	sshConn *ssh.ServerConn,
	newChannel ssh.NewChannel,
	auditConnection *audit.Connection,
) {
	if channelHandler == nil {
		auditConnection.Message(protocol.MessageType_ChannelRequestUnknownType, protocol.PayloadNewChannel{
			ChannelType: newChannel.ChannelType(),
		})
		err := newChannel.Reject(ssh.UnknownChannelType, "no channel channelRequestHandler")
		if err != nil {
			server.logger.DebugF("unable to send channel rejection (%v)", err)
		}
		return
	}
	channelRequestHandler, rejection := channelHandler.OnChannel(
		ctx,
		sshConn.Conn,
		newChannel.ChannelType(),
		newChannel.ExtraData(),
	)
	if rejection != nil {
		err := newChannel.Reject(rejection.RejectionReason, rejection.RejectionMessage)
		if err != nil {
			server.logger.DebugF("unable to send channel rejection (%v)", err)
		}
		return
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		server.logger.DebugF("unable to accept channel (%v)", err)
		return
	}
	defer func() {
		err := channel.Close()
		if err != nil {
			server.logger.DebugF("failed to close channel (%v)", err)
		}
	}()

	for req := range requests {
		channelRequest := req
		if channelRequestHandler == nil {
			server.replyRequest(channelRequest, false, []byte(fmt.Sprintf("unknown request type (%s)", req.Type)))
			continue
		}
		response := channelRequestHandler.OnChannelRequest(ctx, sshConn, channel, channelRequest.Type, channelRequest.Payload)
		if response.Success {
			server.replyRequest(channelRequest, true, response.Payload)
		} else {
			server.logger.DebugF("channel request channelRequestHandler failed (%v)", response.Payload)
			server.replyRequest(channelRequest, false, response.Payload)
		}
	}
}

func (server *Server) replyRequest(channelRequest *ssh.Request, success bool, message []byte) {
	if channelRequest.WantReply {
		err := channelRequest.Reply(success, message)
		if err != nil {
			server.logger.DebugF("failed to send reply (%v)", err)
		}
	}
}
