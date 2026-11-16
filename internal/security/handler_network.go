package security

import (
	"context"
	"sync"

    auth2 "go.containerssh.io/containerssh/auth"
    config2 "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/auth"
    "go.containerssh.io/containerssh/internal/sshserver"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/metadata"
)

type networkHandler struct {
	config  config2.SecurityConfig
	backend sshserver.NetworkConnectionHandler
	logger  log.Logger
}

func (n *networkHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	return n.backend.OnAuthKeyboardInteractive(
		meta,
		challenge,
	)
}

func (n *networkHandler) OnShutdown(shutdownContext context.Context) {
	n.backend.OnShutdown(shutdownContext)
}

func (n *networkHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, password []byte) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return n.backend.OnAuthPassword(meta, password)
}

func (n *networkHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, pubKey auth2.PublicKey) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return n.backend.OnAuthPubKey(meta, pubKey)
}

func (n *networkHandler) NoneAuthEnabled() bool {
	return n.backend.NoneAuthEnabled()
}

func (n *networkHandler) OnAuthNone(meta metadata.ConnectionAuthPendingMetadata) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return n.backend.OnAuthNone(meta)
}

func (n *networkHandler) OnAuthGSSAPI(meta metadata.ConnectionMetadata) auth.GSSAPIServer {
	return n.backend.OnAuthGSSAPI(meta)
}

func (n *networkHandler) OnHandshakeFailed(meta metadata.ConnectionMetadata, reason error) {
	n.backend.OnHandshakeFailed(meta, reason)
}

func (n *networkHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	backend, _, failureReason := n.backend.OnHandshakeSuccess(meta)
	if failureReason != nil {
		return nil, meta, failureReason
	}
	return &sshConnectionHandler{
		config:  n.config,
		backend: backend,
		lock:    &sync.Mutex{},
		logger:  n.logger,
	}, meta, nil
}

func (n *networkHandler) OnDisconnect() {
	n.backend.OnDisconnect()
}
