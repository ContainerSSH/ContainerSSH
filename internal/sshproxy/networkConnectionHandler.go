package sshproxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"

	"golang.org/x/crypto/ssh"

	"github.com/containerssh/libcontainerssh/internal/sshserver"
)

type networkConnectionHandler struct {
	lock                  *sync.Mutex
	wg                    *sync.WaitGroup
	client                net.TCPAddr
	connectionID          string
	config                config.SSHProxyConfig
	logger                log.Logger
	backendRequestsMetric metrics.SimpleCounter
	backendFailuresMetric metrics.SimpleCounter
	tcpConn               net.Conn
	disconnected          bool
	privateKey            ssh.Signer
	done                  bool
}

func (s *networkConnectionHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, _ []byte) (
	_ sshserver.AuthResponse,
	_ metadata.ConnectionAuthenticatedMetadata,
	_ error,
) {
	return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf(
		"ssh proxy does not support authentication",
	)
}

func (s *networkConnectionHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, _ auth2.PublicKey) (
	sshserver.AuthResponse,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf(
		"ssh proxy does not support authentication",
	)
}

func (s *networkConnectionHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	_ func(
		_ string,
		_ sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf(
		"ssh proxy does not support authentication",
	)
}

func (s *networkConnectionHandler) OnAuthGSSAPI(_ metadata.ConnectionMetadata) auth.GSSAPIServer {
	return nil
}

func (s *networkConnectionHandler) OnHandshakeFailed(_ metadata.ConnectionMetadata, _ error) {}

func (s *networkConnectionHandler) OnHandshakeSuccess(
	meta metadata.ConnectionAuthenticatedMetadata,
) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.disconnected {
		return nil, meta, message.NewMessage(
			message.ESSHProxyDisconnected,
			"could not connect to backend because the user already disconnected",
		)
	}
	sshConn, newChannels, requests, err := s.createBackendSSHConnection(meta.Username)
	if err != nil {
		return nil, meta, err
	}

	connectionHandler := &sshConnectionHandler{
		networkHandler: s,
		sshConn:        sshConn,
		logger:         s.logger,
	}
	go connectionHandler.handleChannels(newChannels)
	go connectionHandler.handleRequests(requests)

	return connectionHandler, meta, nil
}

func (s *networkConnectionHandler) createBackendSSHConnection(username string) (
	ssh.Conn,
	<-chan ssh.NewChannel,
	<-chan *ssh.Request,
	error,
) {
	s.backendRequestsMetric.Increment()
	target := fmt.Sprintf("%s:%d", s.config.Server, s.config.Port)
	tcpConn, err := s.createBackendTCPConnection(username, target)
	if err != nil {
		return nil, nil, nil, err
	}
	s.tcpConn = tcpConn

	sshClientConfig := s.createClientConfig(username)

	sshConn, newChannels, requests, err := ssh.NewClientConn(s.tcpConn, target, sshClientConfig)
	if err != nil {
		s.backendFailuresMetric.Increment(metrics.Label("failure", "handshake"))
		return nil, nil, nil, message.WrapUser(
			err,
			message.ESSHProxyBackendHandshakeFailed,
			"SSH service is currently unavailable.",
			"Failed to authenticate with the backend.",
		).Label("backend", target)
	}

	return sshConn, newChannels, requests, nil
}

func (s *networkConnectionHandler) createClientConfig(username string) *ssh.ClientConfig {
	if !s.config.UsernamePassThrough {
		username = s.config.Username
	}

	authMethods := []ssh.AuthMethod{
		ssh.Password(s.config.Password),
	}

	if s.privateKey != nil {
		authMethods = append(
			authMethods, ssh.PublicKeys(
				s.privateKey,
			),
		)
	}
	sshClientConfig := &ssh.ClientConfig{
		Config: ssh.Config{
			KeyExchanges: s.config.KexAlgorithms.StringList(),
			Ciphers:      s.config.Ciphers.StringList(),
			MACs:         s.config.MACs.StringList(),
		},
		User: username,
		Auth: authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			fingerprint := ssh.FingerprintSHA256(key)
			for _, fp := range s.config.AllowedHostKeyFingerprints {
				if fingerprint == fp {
					return nil
				}
			}
			err := message.UserMessage(
				message.ESSHProxyInvalidFingerprint,
				"SSH service currently unavailable",
				"invalid host key fingerprint: %s",
				fingerprint,
			).Label("fingerprint", fingerprint)
			s.logger.Error(err)
			return err
		},
		ClientVersion:     s.config.ClientVersion.String(),
		HostKeyAlgorithms: s.config.HostKeyAlgorithms.StringList(),
		Timeout:           s.config.Timeout,
	}
	return sshClientConfig
}

func (s *networkConnectionHandler) createBackendTCPConnection(
	_ string,
	target string,
) (net.Conn, error) {
	s.logger.Debug(message.NewMessage(message.MSSHProxyConnecting, "Connecting to backend server %s", target))
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancelFunc()
	var networkConnection net.Conn
	var lastError error
loop:
	for {
		networkConnection, lastError = net.Dial("tcp", target)
		if lastError == nil {
			return networkConnection, nil
		}
		s.backendFailuresMetric.Increment(metrics.Label("failure", "tcp"))
		s.logger.Debug(
			message.WrapUser(
				lastError,
				message.ESSHProxyBackendConnectionFailed,
				"service currently unavailable",
				"connection to SSH backend failed, retrying in 10 seconds",
			),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.WrapUser(
		lastError,
		message.ESSHProxyBackendConnectionFailed,
		"service currently unavailable",
		"connection to SSH backend failed, giving up",
	)
	s.logger.Error(err)
	return nil, err
}

func (s *networkConnectionHandler) OnDisconnect() {
	s.logger.Debug(
		message.NewMessage(
			message.MSSHProxyDisconnected,
			"Client disconnected, waiting for network connection lock...",
		),
	)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.logger.Debug(
		message.NewMessage(
			message.MSSHProxyDisconnected,
			"Client disconnected, waiting for all sessions to terminate...",
		),
	)
	s.wg.Wait()
	s.done = true
	s.disconnected = true
	if s.tcpConn != nil {
		s.logger.Debug(message.NewMessage(message.MSSHProxyBackendDisconnecting, "Disconnecting backend connection..."))
		if err := s.tcpConn.Close(); err != nil {
			s.logger.Debug(
				message.Wrap(
					err,
					message.MSSHProxyBackendDisconnectFailed, "Failed to disconnect backend connection.",
				),
			)
		} else {
			s.logger.Debug(message.NewMessage(message.MSSHProxyBackendDisconnected, "Backend connection disconnected."))
		}
	} else {
		s.logger.Debug(
			message.NewMessage(
				message.MSSHProxyBackendDisconnected,
				"Backend connection already disconnected.",
			),
		)
	}
}

func (s *networkConnectionHandler) OnShutdown(_ context.Context) {}
