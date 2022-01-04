package sshserver

import (
	"bytes"
	"context"
	"fmt"
	"net"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/internal/auth"
)

type testAuthenticationNetworkHandler struct {
	rootHandler *testAuthenticationHandler
	backend     NetworkConnectionHandler
}

func (t *testAuthenticationNetworkHandler) OnAuthKeyboardInteractive(
	username string,
	challenge func(
		instruction string,
		questions KeyboardInteractiveQuestions,
	) (answers KeyboardInteractiveAnswers, err error),
	_ string,
) (response AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	var foundUser *TestUser
	for _, user := range t.rootHandler.users {
		if user.Username() == username {
			foundUser = user
			break
		}
	}
	if foundUser == nil {
		return AuthResponseFailure, nil, fmt.Errorf("user not found")
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
		return AuthResponseFailure, nil, err
	}
	for question, expectedAnswer := range foundUser.keyboardInteractive {
		answerText, err := answers.GetByQuestionText(question)
		if err != nil {
			return AuthResponseFailure, nil, err
		}
		if answerText != expectedAnswer {
			return AuthResponseFailure, nil, fmt.Errorf("invalid response")
		}
	}
	return AuthResponseSuccess, nil, nil
}

func (t *testAuthenticationNetworkHandler) OnDisconnect() {
	t.backend.OnDisconnect()
}

func (t *testAuthenticationNetworkHandler) OnShutdown(shutdownContext context.Context) {
	t.backend.OnShutdown(shutdownContext)
}

func (t *testAuthenticationNetworkHandler) OnAuthPassword(username string, password []byte, clientVersion string) (response AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == username && user.password == string(password) {
			return AuthResponseSuccess, nil, nil
		}
	}
	return AuthResponseFailure, nil, ErrAuthenticationFailed
}

func (t *testAuthenticationNetworkHandler) OnAuthPubKey(username string, pubKey string, clientVersion string) (response AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == username {
			for _, authorizedKey := range user.authorizedKeys {
				if pubKey == authorizedKey {
					return AuthResponseSuccess, nil, nil
				}
			}
		}
	}
	return AuthResponseFailure, nil, ErrAuthenticationFailed
}

// NOTE: This is a dummy implementation to test the plumbing, by no means how
// GSSAPI is supposed to work :)
type gssApiServer struct {
	username string
	success bool
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
	return fmt.Errorf("Invalid username")
}

func (s *gssApiServer) DeleteSecContext() error {
	return nil
}

func (s *gssApiServer) AllowLogin(username string) error {
	return nil
}

func (s *gssApiServer) Error() error {
	return nil
}

func (s *gssApiServer) Success() bool {
	return s.success
}

func (s *gssApiServer) Metadata() *auth2.ConnectionMetadata {
	return nil
}

func (s *gssApiServer) OnDisconnect() {
}

func (s *gssApiServer) Password(_ string, _ []byte, _ string, _ net.IP) auth.AuthenticationContext {
	s.success = false
	return s
}

func (s *gssApiServer) PubKey(_ string, _ string, _ string, _ net.IP) auth.AuthenticationContext {
	s.success = false
	return s
}

func (s *gssApiServer) GSSAPIConfig(connectionId string, addr net.IP) auth.GSSAPIServer {
	return s
}

func (t *testAuthenticationNetworkHandler) OnAuthGSSAPI() auth.GSSAPIServer {
	return &gssApiServer{}
}

func (t *testAuthenticationNetworkHandler) OnHandshakeFailed(err error) {
	t.backend.OnHandshakeFailed(err)
}

func (t *testAuthenticationNetworkHandler) OnHandshakeSuccess(username string, clientVersion string, metadata *auth2.ConnectionMetadata) (
	connection SSHConnectionHandler,
	failureReason error,
) {
	return t.backend.OnHandshakeSuccess(username, clientVersion, metadata)
}
