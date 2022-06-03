package authintegration

import (
	"context"
	"net"

    auth2 "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/internal/auth"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
)

// Behavior dictates how when the authentication requests are passed to the backends.
type Behavior int

const (
	// BehaviorNoPassthrough means that the authentication integration will never call the backend for authentication.
	BehaviorNoPassthrough Behavior = iota
	// BehaviorPassthroughOnFailure will call the backend if the authentication server returned a failure.
	BehaviorPassthroughOnFailure Behavior = iota
	// BehaviorPassthroughOnSuccess will call the backend if the authentication server returned a success.
	BehaviorPassthroughOnSuccess Behavior = iota
	// BehaviorPassthroughOnUnavailable will call the backend if the authentication server is not available.
	BehaviorPassthroughOnUnavailable Behavior = iota
)

func (behavior Behavior) validate() bool {
	switch behavior {
	case BehaviorNoPassthrough:
		return true
	case BehaviorPassthroughOnFailure:
		return true
	case BehaviorPassthroughOnSuccess:
		return true
	case BehaviorPassthroughOnUnavailable:
		return true
	default:
		return false
	}
}

type handler struct {
	backend                          sshserver.Handler
	passwordAuthenticator            auth.PasswordAuthenticator
	publicKeyAuthenticator           auth.PublicKeyAuthenticator
	gssapiAuthenticator              auth.GSSAPIAuthenticator
	keyboardInteractiveAuthenticator auth.KeyboardInteractiveAuthenticator
	authorizationProvider            auth.AuthzProvider
	behavior                         Behavior
}

func (h *handler) OnReady() error {
	if h.backend != nil {
		return h.backend.OnReady()
	}
	return nil
}

func (h *handler) OnShutdown(shutdownContext context.Context) {
	if h.backend != nil {
		h.backend.OnShutdown(shutdownContext)
	}
}

func (h *handler) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler,
	metadata.ConnectionMetadata,
	error,
) {
	var backend sshserver.NetworkConnectionHandler = nil
	var err error
	if h.backend != nil {
		backend, meta, err = h.backend.OnNetworkConnection(meta)
		if err != nil {
			return nil, meta, err
		}
	}
	return &networkConnectionHandler{
		connectionID:                     meta.ConnectionID,
		ip:                               meta.RemoteAddress.IP,
		backend:                          backend,
		behavior:                         h.behavior,
		passwordAuthenticator:            h.passwordAuthenticator,
		publicKeyAuthenticator:           h.publicKeyAuthenticator,
		gssapiAuthenticator:              h.gssapiAuthenticator,
		keyboardInteractiveAuthenticator: h.keyboardInteractiveAuthenticator,
		authorizationProvider:            h.authorizationProvider,
	}, meta, nil
}

type networkConnectionHandler struct {
	backend                          sshserver.NetworkConnectionHandler
	ip                               net.IP
	connectionID                     string
	behavior                         Behavior
	authContext                      auth.AuthenticationContext
	passwordAuthenticator            auth.PasswordAuthenticator
	publicKeyAuthenticator           auth.PublicKeyAuthenticator
	gssapiAuthenticator              auth.GSSAPIAuthenticator
	keyboardInteractiveAuthenticator auth.KeyboardInteractiveAuthenticator
	authorizationProvider            auth.AuthzProvider
}

func (h *networkConnectionHandler) OnShutdown(shutdownContext context.Context) {
	h.backend.OnShutdown(shutdownContext)
}

func (h *networkConnectionHandler) OnAuthPassword(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) (response sshserver.AuthResponse, returnMeta metadata.ConnectionAuthenticatedMetadata, reason error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	if h.passwordAuthenticator == nil {
		return sshserver.AuthResponseUnavailable, meta.AuthFailed(), message.UserMessage(
			message.ESSHAuthUnavailable,
			"This authentication method is currently unavailable.",
			"Password authentication is disabled.",
		)
	}
	authContext := h.passwordAuthenticator.Password(meta, password)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthPassword(meta, password)
			}
			return sshserver.AuthResponseUnavailable, meta.AuthFailed(), authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthPassword(meta, password)
		}
		return sshserver.AuthResponseFailure, meta.AuthFailed(), nil
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthPassword(meta, password)
	} else {
		return sshserver.AuthResponseSuccess, authContext.Metadata(), nil
	}
}

func (h *networkConnectionHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	pubKey auth2.PublicKey,
) (response sshserver.AuthResponse, meta2 metadata.ConnectionAuthenticatedMetadata, reason error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	if h.publicKeyAuthenticator == nil {
		return sshserver.AuthResponseUnavailable, meta.AuthFailed(), message.UserMessage(
			message.ESSHAuthUnavailable,
			"This authentication method is currently unavailable.",
			"Public key authentication is disabled.",
		)
	}
	authContext := h.publicKeyAuthenticator.PubKey(meta, pubKey)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthPubKey(meta, pubKey)
			}
			return sshserver.AuthResponseUnavailable, authContext.Metadata(), authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthPubKey(meta, pubKey)
		}
		return sshserver.AuthResponseFailure, authContext.Metadata(), authContext.Error()
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthPubKey(meta, pubKey)
	}
	return sshserver.AuthResponseSuccess, authContext.Metadata(), authContext.Error()
}

func (h *networkConnectionHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	if h.keyboardInteractiveAuthenticator == nil {
		return sshserver.AuthResponseUnavailable, meta.AuthFailed(), message.UserMessage(
			message.ESSHAuthUnavailable,
			"This authentication method is currently unavailable.",
			"Keyboard-interactive authentication is disabled.",
		)
	}
	authContext := h.keyboardInteractiveAuthenticator.KeyboardInteractive(
		meta,
		func(instruction string, questions auth.KeyboardInteractiveQuestions) (
			answers auth.KeyboardInteractiveAnswers,
			err error,
		) {
			q := make(sshserver.KeyboardInteractiveQuestions, len(questions))
			for i, question := range questions {
				q[i] = sshserver.KeyboardInteractiveQuestion(question)
			}
			a, err := challenge(instruction, q)
			if err != nil {
				return auth.KeyboardInteractiveAnswers{}, err
			}
			answers = auth.KeyboardInteractiveAnswers{
				Answers: make(map[string]string, len(questions)),
			}
			for i, question := range questions {
				ans, err := a.Get(q[i])
				if err != nil {
					ans = ""
				}
				answers.Answers[question.ID] = ans
			}
			return answers, nil
		},
	)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthKeyboardInteractive(meta, challenge)
			}
			return sshserver.AuthResponseUnavailable, authContext.Metadata(), authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthKeyboardInteractive(meta, challenge)
		}
		return sshserver.AuthResponseFailure, authContext.Metadata(), authContext.Error()
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthKeyboardInteractive(meta, challenge)
	}
	return sshserver.AuthResponseSuccess, authContext.Metadata(), authContext.Error()
}

func (h *networkConnectionHandler) OnAuthGSSAPI(meta metadata.ConnectionMetadata) auth.GSSAPIServer {
	if h.gssapiAuthenticator == nil {
		return nil
	}
	return h.gssapiAuthenticator.GSSAPI(meta)
}

func (h *networkConnectionHandler) OnHandshakeFailed(meta metadata.ConnectionMetadata, reason error) {
	h.backend.OnHandshakeFailed(meta, reason)
}

func (h *networkConnectionHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	sshserver.SSHConnectionHandler,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return h.backend.OnHandshakeSuccess(meta)
}

func (h *networkConnectionHandler) OnDisconnect() {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	h.backend.OnDisconnect()
}
