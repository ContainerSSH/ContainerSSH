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

	authHandler := networkConnectionHandler{
		connectionID:                     meta.ConnectionID,
		ip:                               meta.RemoteAddress.IP,
		backend:                          backend,
		behavior:                         h.behavior,
		passwordAuthenticator:            h.passwordAuthenticator,
		publicKeyAuthenticator:           h.publicKeyAuthenticator,
		gssapiAuthenticator:              h.gssapiAuthenticator,
		keyboardInteractiveAuthenticator: h.keyboardInteractiveAuthenticator,
		authorizationProvider:            h.authorizationProvider,
	}

	if h.authorizationProvider != nil {
		// We inject the authz handler before the normal authentication handler in the chain as we need the authenticated metadata the handler returns.
		// Authentications request will first hit the authz handler which will pass it through to the authHandler, once it returns we can perform authorization.
		authzHandler := authzNetworkConnectionHandler{
			connectionID:          meta.ConnectionID,
			ip:                    meta.RemoteAddress.IP,
			authorizationProvider: h.authorizationProvider,
			backend:               &authHandler,
		}
		return &authzHandler, meta, nil
	}
	return &authHandler, meta, nil
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

type authzNetworkConnectionHandler struct {
	backend               sshserver.NetworkConnectionHandler
	ip                    net.IP
	connectionID          string
	authorizationProvider auth.AuthzProvider
}

// genericAuthorization is a helper function that takes the response of an authentication call (e.g. OnAuthPassword) and performs authorization.
func (a *authzNetworkConnectionHandler) genericAuthorization(
	meta metadata.ConnectionAuthPendingMetadata,
	authResponse sshserver.AuthResponse,
	authenticatedMeta metadata.ConnectionAuthenticatedMetadata,
	err error,
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	if authResponse != sshserver.AuthResponseSuccess {
		return authResponse, authenticatedMeta, err
	}

	authzResponse := a.authorizationProvider.Authorize(authenticatedMeta)
	if authzResponse.Success() {
		return sshserver.AuthResponseSuccess, authzResponse.Metadata(), err
	}
	return sshserver.AuthResponseFailure, authzResponse.Metadata(), authzResponse.Error()
}

// OnAuthPassword is called when a user attempts a password authentication. The implementation must always supply
// AuthResponse and may supply error as a reason description.
func (a *authzNetworkConnectionHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, password []byte) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	authResponse, authenticatedMeta, err := a.backend.OnAuthPassword(meta, password)
	return a.genericAuthorization(meta, authResponse, authenticatedMeta, err)
}

// OnAuthPubKey is called when a user attempts a pubkey authentication. The implementation must always supply
// AuthResponse and may supply error as a reason description. The pubKey parameter is an SSH key in
// the form of "ssh-rsa KEY HERE".
func (a *authzNetworkConnectionHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, pubKey auth2.PublicKey) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	authResponse, authenticatedMeta, err := a.backend.OnAuthPubKey(meta, pubKey)
	return a.genericAuthorization(meta, authResponse, authenticatedMeta, err)
}

// OnAuthKeyboardInteractive is a callback for interactive authentication. The implementer will be passed a callback
// function that can be used to issue challenges to the user. These challenges can, but do not have to contain
// questions.
func (a *authzNetworkConnectionHandler) OnAuthKeyboardInteractive(meta metadata.ConnectionAuthPendingMetadata, challenge func(
	instruction string,
	questions sshserver.KeyboardInteractiveQuestions) (answers sshserver.KeyboardInteractiveAnswers, err error)) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	authResponse, authenticatedMeta, err := a.backend.OnAuthKeyboardInteractive(meta, challenge)
	return a.genericAuthorization(meta, authResponse, authenticatedMeta, err)
}

// OnAuthGSSAPI returns a GSSAPIServer which can perform a GSSAPI authentication.
func (a *authzNetworkConnectionHandler) OnAuthGSSAPI(metadata metadata.ConnectionMetadata) auth.GSSAPIServer {
	gssApiServer := a.backend.OnAuthGSSAPI(metadata)
	authzGssApiServer := authzGssApiServer{
		backend: gssApiServer,
	}
	return &authzGssApiServer
}

// OnHandshakeFailed is called when the SSH handshake failed. This method is also called after an authentication
// failure. After this method is the connection will be closed and the OnDisconnect method will be
// called.
func (a *authzNetworkConnectionHandler) OnHandshakeFailed(metadata metadata.ConnectionMetadata, reason error) {
	a.backend.OnHandshakeFailed(metadata, reason)
}

// OnHandshakeSuccess is called when the SSH handshake was successful. It returns metadata to process
// requests, or failureReason to indicate that a backend error has happened. In this case, the
// metadata will be closed and OnDisconnect will be called.
func (a *authzNetworkConnectionHandler) OnHandshakeSuccess(metadata metadata.ConnectionAuthenticatedMetadata) (connection sshserver.SSHConnectionHandler, meta metadata.ConnectionAuthenticatedMetadata, failureReason error) {
	return a.backend.OnHandshakeSuccess(metadata)
}

// OnDisconnect is called when the network connection is closed.
func (a *authzNetworkConnectionHandler) OnDisconnect() {
	a.backend.OnDisconnect()
}

// OnShutdown is called when a shutdown of the SSH server is desired. The shutdownContext is passed as a deadline
// for the shutdown, after which the server should abort all running connections and return as fast as
// possible.
func (a *authzNetworkConnectionHandler) OnShutdown(shutdownContext context.Context) {
	a.backend.OnShutdown(shutdownContext)
}

type authzGssApiServer struct {
	backend               auth.GSSAPIServer
	authorizationProvider auth.AuthzProvider
	authzResponse         auth.AuthorizationResponse
}

// Success must return true or false of the authentication was successful / unsuccessful.
func (g *authzGssApiServer) Success() bool {
	backendResponse := g.backend.Success()
	if !backendResponse || g.authzResponse == nil {
		return false
	}
	return g.authzResponse.Success()
}

// Error returns the error that happened during the authentication.
func (g *authzGssApiServer) Error() error {
	backendErr := g.backend.Error()
	if backendErr != nil || g.authzResponse == nil {
		return backendErr
	}
	return g.authzResponse.Error()
}

// AcceptSecContext is the GSSAPI function to verify the tokens.
func (g *authzGssApiServer) AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error) {
	return g.backend.AcceptSecContext(token)
}

// VerifyMIC is the GSSAPI function to verify the MIC (Message Integrity Code).
func (g *authzGssApiServer) VerifyMIC(micField []byte, micToken []byte) error {
	return g.backend.VerifyMIC(micField, micToken)
}

// DeleteSecContext is the GSSAPI function to free all resources bound as part of an authentication attempt.
func (g *authzGssApiServer) DeleteSecContext() error {
	return g.backend.DeleteSecContext()
}

// AllowLogin is the authorization function. The username parameter
// specifies the user that the authenticated user is trying to log in
// as. Note! This is different from the gossh AllowLogin function in
// which the username field is the authenticated username.
func (g *authzGssApiServer) AllowLogin(username string, meta metadata.ConnectionAuthPendingMetadata) (metadata.ConnectionAuthenticatedMetadata, error) {
	authenticatedMetadata, err := g.backend.AllowLogin(username, meta)
	if err != nil {
		return authenticatedMetadata, err
	}

	authzResponse := g.authorizationProvider.Authorize(authenticatedMetadata)
	g.authzResponse = authzResponse
	return authzResponse.Metadata(), authzResponse.Error()
}
