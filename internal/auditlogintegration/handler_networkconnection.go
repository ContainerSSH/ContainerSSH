package auditlogintegration

import (
	"context"

	"go.containerssh.io/libcontainerssh/auditlog/message"
	publicAuth "go.containerssh.io/libcontainerssh/auth"
	"go.containerssh.io/libcontainerssh/internal/auditlog"
	internalAuth "go.containerssh.io/libcontainerssh/internal/auth"
	"go.containerssh.io/libcontainerssh/internal/sshserver"
	"go.containerssh.io/libcontainerssh/metadata"
)

type networkConnectionHandler struct {
	backend sshserver.NetworkConnectionHandler
	audit   auditlog.Connection
}

func (n *networkConnectionHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (response sshserver.AuthResponse, metadata metadata.ConnectionAuthenticatedMetadata, reason error) {
	return n.backend.OnAuthKeyboardInteractive(
		meta,
		func(
			instruction string,
			questions sshserver.KeyboardInteractiveQuestions,
		) (answers sshserver.KeyboardInteractiveAnswers, err error) {
			var auditQuestions []message.KeyboardInteractiveQuestion
			for _, q := range questions {
				auditQuestions = append(
					auditQuestions, message.KeyboardInteractiveQuestion{
						Question: q.Question,
						Echo:     q.EchoResponse,
					},
				)
			}
			n.audit.OnAuthKeyboardInteractiveChallenge(meta.Username, instruction, auditQuestions)
			answers, err = challenge(instruction, questions)
			if err != nil {
				return answers, err
			}
			var auditAnswers []message.KeyboardInteractiveAnswer
			for _, q := range questions {
				a, err := answers.Get(q)
				if err != nil {
					return answers, err
				}
				auditAnswers = append(auditAnswers, message.KeyboardInteractiveAnswer{
					Question: q.Question,
					Answer:   a,
				})
			}
			n.audit.OnAuthKeyboardInteractiveAnswer(meta.Username, auditAnswers)
			return answers, err
		},
	)
}

func (n *networkConnectionHandler) OnShutdown(shutdownContext context.Context) {
	n.backend.OnShutdown(shutdownContext)
}

func (n *networkConnectionHandler) OnAuthPassword(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) (response sshserver.AuthResponse, authenticatedMetadata metadata.ConnectionAuthenticatedMetadata, reason error) {
	n.audit.OnAuthPassword(meta.Username, password)
	response, authenticatedMetadata, reason = n.backend.OnAuthPassword(meta, password)
	switch response {
	case sshserver.AuthResponseSuccess:
		// TODO add authenticated username
		n.audit.OnAuthPasswordSuccess(meta.Username, password)
	case sshserver.AuthResponseFailure:
		// TODO add authenticated username
		n.audit.OnAuthPasswordFailed(meta.Username, password)
	case sshserver.AuthResponseUnavailable:
		if reason != nil {
			n.audit.OnAuthPasswordBackendError(meta.Username, password, reason.Error())
		} else {
			n.audit.OnAuthPasswordBackendError(meta.Username, password, "")
		}
	}
	return response, authenticatedMetadata, reason
}

func (n *networkConnectionHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	pubKey publicAuth.PublicKey,
) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	// TODO add authenticated username
	n.audit.OnAuthPubKey(meta.Username, pubKey.PublicKey)
	response, authMeta, reason := n.backend.OnAuthPubKey(meta, pubKey)
	switch response {
	case sshserver.AuthResponseSuccess:
		n.audit.OnAuthPubKeySuccess(authMeta.Username, pubKey.PublicKey)
	case sshserver.AuthResponseFailure:
		n.audit.OnAuthPubKeyFailed(authMeta.Username, pubKey.PublicKey)
	case sshserver.AuthResponseUnavailable:
		if reason != nil {
			n.audit.OnAuthPubKeyBackendError(authMeta.Username, pubKey.PublicKey, reason.Error())
		} else {
			n.audit.OnAuthPubKeyBackendError(authMeta.Username, pubKey.PublicKey, "")
		}
	}
	return response, authMeta, reason
}

func (n *networkConnectionHandler) OnAuthGSSAPI(meta metadata.ConnectionMetadata) internalAuth.GSSAPIServer {
	// TODO add audit logging
	return n.backend.OnAuthGSSAPI(meta)
}

func (n *networkConnectionHandler) OnHandshakeFailed(meta metadata.ConnectionMetadata, reason error) {
	n.backend.OnHandshakeFailed(meta, reason)
	n.audit.OnHandshakeFailed(reason.Error())
}

func (n *networkConnectionHandler) OnHandshakeSuccess(
	meta metadata.ConnectionAuthenticatedMetadata,
) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	// TODO log authenticated username
	n.audit.OnHandshakeSuccessful(meta.Username)
	backend, meta, err := n.backend.OnHandshakeSuccess(meta)
	if err != nil {
		return nil, meta, err
	}
	return &sshConnectionHandler{
		backend: backend,
		audit:   n.audit,
	}, meta, nil
}

func (n *networkConnectionHandler) OnDisconnect() {
	n.audit.OnDisconnect()
	n.backend.OnDisconnect()
}
