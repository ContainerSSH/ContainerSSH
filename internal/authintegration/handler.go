package authintegration

import (
	"context"
	"net"

	auth2 "github.com/containerssh/libcontainerssh/auth"

	"github.com/containerssh/libcontainerssh/internal/sshserver"

	"github.com/containerssh/libcontainerssh/internal/auth"
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
	backend    sshserver.Handler
	authClient auth.Client
	behavior   Behavior
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

func (h *handler) OnNetworkConnection(client net.TCPAddr, connectionID string) (sshserver.NetworkConnectionHandler, error) {
	var backend sshserver.NetworkConnectionHandler = nil
	var err error
	if h.backend != nil {
		backend, err = h.backend.OnNetworkConnection(client, connectionID)
		if err != nil {
			return nil, err
		}
	}
	return &networkConnectionHandler{
		connectionID: connectionID,
		ip:           client.IP,
		authClient:   h.authClient,
		backend:      backend,
		behavior:     h.behavior,
	}, nil
}

type networkConnectionHandler struct {
	backend      sshserver.NetworkConnectionHandler
	authClient   auth.Client
	ip           net.IP
	connectionID string
	behavior     Behavior
	authContext  auth.AuthenticationContext
}

func (h *networkConnectionHandler) OnShutdown(shutdownContext context.Context) {
	h.backend.OnShutdown(shutdownContext)
}

func (h *networkConnectionHandler) OnAuthPassword(username string, password []byte, clientVersion string) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	authContext := h.authClient.Password(username, password, h.connectionID, h.ip)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthPassword(username, password, clientVersion)
			}
			return sshserver.AuthResponseUnavailable, nil, authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthPassword(username, password, clientVersion)
		}
		return sshserver.AuthResponseFailure, nil, nil
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthPassword(username, password, clientVersion)
	} else {
		return sshserver.AuthResponseSuccess, authContext.Metadata(), nil
	}
}

func (h *networkConnectionHandler) OnAuthPubKey(username string, pubKey string, clientVersion string) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	authContext := h.authClient.PubKey(username, pubKey, h.connectionID, h.ip)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthPubKey(username, pubKey, clientVersion)
			}
			return sshserver.AuthResponseUnavailable, authContext.Metadata(), authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthPubKey(username, pubKey, clientVersion)
		}
		return sshserver.AuthResponseFailure, authContext.Metadata(), authContext.Error()
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthPubKey(username, pubKey, clientVersion)
	}
	return sshserver.AuthResponseSuccess, authContext.Metadata(), authContext.Error()
}

func (h *networkConnectionHandler) OnAuthKeyboardInteractive(
	username string,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
	clientVersion string,
) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	authContext := h.authClient.KeyboardInteractive(username,
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
		}, h.connectionID, h.ip)
	h.authContext = authContext
	if !authContext.Success() {
		if authContext.Error() != nil {
			if h.behavior == BehaviorPassthroughOnUnavailable {
				return h.backend.OnAuthKeyboardInteractive(username, challenge, clientVersion)
			}
			return sshserver.AuthResponseUnavailable, authContext.Metadata(), authContext.Error()
		}
		if h.behavior == BehaviorPassthroughOnFailure {
			return h.backend.OnAuthKeyboardInteractive(username, challenge, clientVersion)
		}
		return sshserver.AuthResponseFailure, authContext.Metadata(), authContext.Error()
	}
	if h.behavior == BehaviorPassthroughOnSuccess {
		return h.backend.OnAuthKeyboardInteractive(username, challenge, clientVersion)
	}
	return sshserver.AuthResponseSuccess, authContext.Metadata(), authContext.Error()
}

func (h *networkConnectionHandler) OnAuthGSSAPI() auth.GSSAPIServer {
	return h.authClient.GSSAPIConfig(h.connectionID, h.ip)
}

func (h *networkConnectionHandler) OnHandshakeFailed(reason error) {
	h.backend.OnHandshakeFailed(reason)
}

func (h *networkConnectionHandler) OnHandshakeSuccess(username string, clientVersion string, metadata *auth2.ConnectionMetadata) (connection sshserver.SSHConnectionHandler, failureReason error) {
	return h.backend.OnHandshakeSuccess(username, clientVersion, metadata)
}

func (h *networkConnectionHandler) OnDisconnect() {
	if h.authContext != nil {
		h.authContext.OnDisconnect()
	}
	h.backend.OnDisconnect()
}
