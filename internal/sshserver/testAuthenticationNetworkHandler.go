package sshserver

import (
	"context"
	"fmt"
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
) (response AuthResponse, metadata map[string]string, reason error) {
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

	var questions []KeyboardInteractiveQuestion
	for question := range foundUser.keyboardInteractive {
		questions = append(questions, KeyboardInteractiveQuestion{
			ID:           question,
			Question:     question,
			EchoResponse: false,
		})
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

func (t *testAuthenticationNetworkHandler) OnAuthPassword(username string, password []byte, clientVersion string) (response AuthResponse, metadata map[string]string, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == username && user.password == string(password) {
			return AuthResponseSuccess, nil, nil
		}
	}
	return AuthResponseFailure, nil, ErrAuthenticationFailed
}

func (t *testAuthenticationNetworkHandler) OnAuthPubKey(username string, pubKey string, clientVersion string) (response AuthResponse, metadata map[string]string, reason error) {
	for _, user := range t.rootHandler.users {
		if user.username == username {
			for _, authorizedKey := range user.authorizedKeys {
				if pubKey == authorizedKey {
					return AuthResponseSuccess, nil,nil
				}
			}
		}
	}
	return AuthResponseFailure, nil, ErrAuthenticationFailed
}

func (t *testAuthenticationNetworkHandler) OnHandshakeFailed(err error) {
	t.backend.OnHandshakeFailed(err)
}

func (t *testAuthenticationNetworkHandler) OnHandshakeSuccess(username string, clientVersion string, metadata map[string]string) (
	connection SSHConnectionHandler,
	failureReason error,
) {
	return t.backend.OnHandshakeSuccess(username, clientVersion, metadata)
}
