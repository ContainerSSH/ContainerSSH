package auditlogintegration

import (
	"context"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/auditlog/message"
	"github.com/containerssh/libcontainerssh/internal/auditlog"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
)

type networkConnectionHandler struct {
	backend sshserver.NetworkConnectionHandler
	audit   auditlog.Connection
}

func (n *networkConnectionHandler) OnAuthKeyboardInteractive(
	user string,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
	clientVersion string,
) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	return n.backend.OnAuthKeyboardInteractive(
		user,
		func(
			instruction string,
			questions sshserver.KeyboardInteractiveQuestions,
		) (answers sshserver.KeyboardInteractiveAnswers, err error) {
			var auditQuestions []message.KeyboardInteractiveQuestion
			for _, q := range questions {
				auditQuestions = append(auditQuestions, message.KeyboardInteractiveQuestion{
					Question: q.Question,
					Echo:     q.EchoResponse,
				})
			}
			n.audit.OnAuthKeyboardInteractiveChallenge(user, instruction, auditQuestions)
			answers, err = challenge(instruction, questions)
			if err != nil {
				return answers, err
			}
			var auditAnswers []message.KeyboardInteractiveAnswer
			for _, q := range auditQuestions {
				a, err := answers.GetByQuestionText(q.Question)
				if err != nil {
					return answers, err
				}
				auditAnswers = append(auditAnswers, message.KeyboardInteractiveAnswer{
					Question: q.Question,
					Answer:   a,
				})
			}
			n.audit.OnAuthKeyboardInteractiveAnswer(user, auditAnswers)
			return answers, err
		},
		clientVersion,
	)
}

func (n *networkConnectionHandler) OnShutdown(shutdownContext context.Context) {
	n.backend.OnShutdown(shutdownContext)
}

func (n *networkConnectionHandler) OnAuthPassword(
	username string,
	password []byte,
	clientVersion string,
) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	n.audit.OnAuthPassword(username, password)
	response, metadata, reason = n.backend.OnAuthPassword(username, password, clientVersion)
	switch response {
	case sshserver.AuthResponseSuccess:
		n.audit.OnAuthPasswordSuccess(username, password)
	case sshserver.AuthResponseFailure:
		n.audit.OnAuthPasswordFailed(username, password)
	case sshserver.AuthResponseUnavailable:
		if reason != nil {
			n.audit.OnAuthPasswordBackendError(username, password, reason.Error())
		} else {
			n.audit.OnAuthPasswordBackendError(username, password, "")
		}
	}
	return response, metadata, reason
}

func (n *networkConnectionHandler) OnAuthPubKey(
	username string,
	pubKey string,
	clientVersion string,
) (
	response sshserver.AuthResponse,
	metadata *auth2.ConnectionMetadata,
	reason error,
) {
	n.audit.OnAuthPubKey(username, pubKey)
	response, metadata, reason = n.backend.OnAuthPubKey(username, pubKey, clientVersion)
	switch response {
	case sshserver.AuthResponseSuccess:
		n.audit.OnAuthPubKeySuccess(username, pubKey)
	case sshserver.AuthResponseFailure:
		n.audit.OnAuthPubKeyFailed(username, pubKey)
	case sshserver.AuthResponseUnavailable:
		if reason != nil {
			n.audit.OnAuthPubKeyBackendError(username, pubKey, reason.Error())
		} else {
			n.audit.OnAuthPubKeyBackendError(username, pubKey, "")
		}
	}
	return response, metadata, reason
}

func (n *networkConnectionHandler) OnAuthGSSAPI() auth.GSSAPIServer {
	return n.backend.OnAuthGSSAPI()
}

func (n *networkConnectionHandler) OnHandshakeFailed(reason error) {
	n.backend.OnHandshakeFailed(reason)
	n.audit.OnHandshakeFailed(reason.Error())
}

func (n *networkConnectionHandler) OnHandshakeSuccess(
	username string,
	clientVersion string,
	metadata *auth2.ConnectionMetadata,
) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	n.audit.OnHandshakeSuccessful(username)
	backend, err := n.backend.OnHandshakeSuccess(username, clientVersion, metadata)
	if err != nil {
		return nil, err
	}
	return &sshConnectionHandler{
		backend: backend,
		audit:   n.audit,
	}, nil
}

func (n *networkConnectionHandler) OnDisconnect() {
	n.audit.OnDisconnect()
	n.backend.OnDisconnect()
}
