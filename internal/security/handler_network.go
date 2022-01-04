package security

import (
	"context"
	"sync"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/internal/auth"
	config2 "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

type networkHandler struct {
	config  config2.SecurityConfig
	backend sshserver.NetworkConnectionHandler
	logger  log.Logger
}

func (n *networkHandler) OnAuthKeyboardInteractive(
	user string,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
	clientVersion string,
) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	return n.backend.OnAuthKeyboardInteractive(
		user,
		challenge,
		clientVersion,
	)
}

func (n *networkHandler) OnShutdown(shutdownContext context.Context) {
	n.backend.OnShutdown(shutdownContext)
}

func (n *networkHandler) OnAuthPassword(username string, password []byte, clientVersion string) (
	response sshserver.AuthResponse,
	metadata *auth2.ConnectionMetadata,
	reason error,
) {
	return n.backend.OnAuthPassword(username, password, clientVersion)
}

func (n *networkHandler) OnAuthPubKey(username string, pubKey string, clientVersion string) (
	response sshserver.AuthResponse,
	metadata *auth2.ConnectionMetadata,
	reason error,
) {
	return n.backend.OnAuthPubKey(username, pubKey, clientVersion)
}

func (n *networkHandler) OnAuthGSSAPI() auth.GSSAPIServer {
	return n.backend.OnAuthGSSAPI()
}

func (n *networkHandler) OnHandshakeFailed(reason error) {
	n.backend.OnHandshakeFailed(reason)
}

func (n *networkHandler) OnHandshakeSuccess(username string, clientVersion string, metadata *auth2.ConnectionMetadata) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	backend, failureReason := n.backend.OnHandshakeSuccess(username, clientVersion, metadata)
	if failureReason != nil {
		return nil, failureReason
	}
	return &sshConnectionHandler{
		config:  n.config,
		backend: backend,
		lock:    &sync.Mutex{},
		logger:  n.logger,
	}, nil
}

func (n *networkHandler) OnDisconnect() {
	n.backend.OnDisconnect()
}
