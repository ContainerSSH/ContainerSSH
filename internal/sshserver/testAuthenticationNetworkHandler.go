package sshserver

import (
	"bytes"
	"context"
	"fmt"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/metadata"
)

type testAuthenticationNetworkHandler struct {
	rootHandler *testAuthenticationHandler
	backend     NetworkConnectionHandler
}

func (t *testAuthenticationNetworkHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions KeyboardInteractiveQuestions,
	) (answers KeyboardInteractiveAnswers, err error),
) (response AuthResponse, authenticatedMetadata metadata.ConnectionAuthenticatedMetadata, reason error) {
	var foundUser *TestUser
	for _, user := range t.rootHandler.users {
		if user.Username() == meta.Username {
			foundUser = user
			break
		}
	}
	if foundUser == nil {
		return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, fmt.Errorf("user not found")
	}

	questions := make([]KeyboardInteractiveQuestion, len(foundUser.keyboardInteractive))
	i := 0
	for question := range foundUser.keyboardInteractive {
		questions[i] = KeyboardInteractiveQuestion{
			ID:           question,
			Question:     question,
			EchoResponse: false,
		}
		i++
	}

	answers, err := challenge("", questions)
	if err != nil {
		return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, err
	}
	for question, expectedAnswer := range foundUser.keyboardInteractive {
		answerText, err := answers.GetByQuestionText(question)
		if err != nil {
			return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, err
		}
		if answerText != expectedAnswer {
			return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, fmt.Errorf("invalid response")
		}
	}
	return AuthResponseSuccess, meta.Authenticated(foundUser.username), nil
}

func (t *testAuthenticationNetworkHandler) OnDisconnect() {
	t.backend.OnDisconnect()
}

func (t *testAuthenticationNetworkHandler) OnShutdown(shutdownContext context.Context) {
	t.backend.OnShutdown(shutdownContext)
}

func (t *testAuthenticationNetworkHandler) OnAuthPassword(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) (response AuthResponse, responseMeta metadata.ConnectionAuthenticatedMetadata, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == meta.Username && user.password == string(password) {
			return AuthResponseSuccess, meta.Authenticated(user.username), nil
		}
	}
	return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, ErrAuthenticationFailed
}

func (t *testAuthenticationNetworkHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	pubKey auth2.PublicKey,
) (response AuthResponse, responseMeta metadata.ConnectionAuthenticatedMetadata, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == meta.Username {
			for _, authorizedKey := range user.authorizedKeys {
				if pubKey.PublicKey == authorizedKey {
					return AuthResponseSuccess, meta.Authenticated(user.username), nil
				}
			}
		}
	}
	return AuthResponseFailure, metadata.ConnectionAuthenticatedMetadata{}, ErrAuthenticationFailed
}

// NOTE: This is a dummy implementation to test the plumbing, by no means how
// GSSAPI is supposed to work :)
type gssApiServer struct {
	username string
	success  bool
}

func (s *gssApiServer) AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error) {
	s.username = string(token)
	s.success = true
	return token, s.username, false, nil
}

func (s *gssApiServer) VerifyMIC(micField []byte, micToken []byte) error {
	if bytes.Equal(micField, []byte(s.username)) {
		s.success = true
		return nil
	}
	return fmt.Errorf("invalid username")
}

func (s *gssApiServer) DeleteSecContext() error {
	return nil
}

func (s *gssApiServer) AllowLogin(
	username string,
	meta metadata.ConnectionAuthPendingMetadata,
) (metadata.ConnectionAuthenticatedMetadata, error) {
	if s.success {
		return meta.Authenticated(username), nil
	}
	return meta.AuthFailed(), nil
}

func (s *gssApiServer) Error() error {
	return nil
}

func (s *gssApiServer) Success() bool {
	return s.success
}

func (s *gssApiServer) OnDisconnect() {
}

func (s *gssApiServer) GSSAPI(meta metadata.ConnectionMetadata) auth.GSSAPIServer {
	return s
}

func (t *testAuthenticationNetworkHandler) OnAuthGSSAPI(_ metadata.ConnectionMetadata) auth.GSSAPIServer {
	return &gssApiServer{}
}

func (t *testAuthenticationNetworkHandler) OnHandshakeFailed(meta metadata.ConnectionMetadata, err error) {
	t.backend.OnHandshakeFailed(meta, err)
}

func (t *testAuthenticationNetworkHandler) OnHandshakeSuccess(authenticatedMetadata metadata.ConnectionAuthenticatedMetadata) (
	connection SSHConnectionHandler,
	meta metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	return t.backend.OnHandshakeSuccess(authenticatedMetadata)
}
