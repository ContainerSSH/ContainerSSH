package auth

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

type httpAuthzClient struct {
	backend               Client
	timeout               time.Duration
	httpClient            http.Client
	endpoint              string
	logger                log.Logger
	metrics               metrics.Collector
	backendRequestsMetric metrics.SimpleCounter
	backendFailureMetric  metrics.SimpleCounter
	authSuccessMetric     metrics.GeoCounter
	authFailureMetric     metrics.GeoCounter
}

type authzContext struct {
	client        *httpAuthzClient
	backend       GSSAPIServer
	princUsername string
	connectionID  string
	remoteAddr    net.IP

	success  bool
	err      error
}

func (h authzContext) Success() bool {
	if h.success {
		return true
	}
	return h.backend.Success()
}

func (h authzContext) Error() error {
	if h.err != nil {
		return h.err
	}
	return h.backend.Error()
}

func (h authzContext) Metadata() *auth.ConnectionMetadata {
	return h.backend.Metadata()
}

func (h authzContext) OnDisconnect() {
}

func (client *httpAuthzClient) KeyboardInteractive(
	username string,
	challenge func(instruction string, questions KeyboardInteractiveQuestions) (
		answers KeyboardInteractiveAnswers,
		err error,
	),
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	auth := client.backend.KeyboardInteractive(username, challenge, connectionID, remoteAddr)
	if !auth.Success() {
		return auth
	}

	return client.processAuthzWithRetry(username, username, connectionID, remoteAddr)
}

func (client *httpAuthzClient) Password(
	username string,
	password []byte,
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	auth := client.backend.Password(username, password, connectionID, remoteAddr)
	if !auth.Success() {
		return auth
	}

	return client.processAuthzWithRetry(username, username, connectionID, remoteAddr)
}

func (client *httpAuthzClient) PubKey(
	username string,
	pubKey string,
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	auth := client.backend.PubKey(username, pubKey, connectionID, remoteAddr)
	if !auth.Success() {
		return auth
	}

	return client.processAuthzWithRetry(username, username, connectionID, remoteAddr)
}

func (s *authzContext) AcceptSecContext(token []byte) (outputToken []byte, srcName string, needContinue bool, err error) {
	outputToken, srcName, needContinue, err = s.backend.AcceptSecContext(token)
	s.princUsername = srcName
	return outputToken, srcName, needContinue, err
}


func (s *authzContext) VerifyMIC(micField []byte, micToken []byte) error {
	return s.backend.VerifyMIC(micField, micToken)
}

func (s *authzContext) DeleteSecContext() error {
	return s.backend.DeleteSecContext()
}


func (s *authzContext) AllowLogin(username string) error {
	err := s.backend.AllowLogin(username)
	if err != nil {
		return err
	}

	authz := s.client.processAuthzWithRetry(s.princUsername, username, s.connectionID, s.remoteAddr)
	if authz.Error() != nil {
		return authz.Error()
	}
	if !authz.Success() {
		return message.NewMessage(
			message.EAuthzFailed,
			"Authorization failed for principal %s trying to log in as %s",
			s.princUsername,
			username,
		)
	}
	return nil
}

func (client *httpAuthzClient) GSSAPIConfig(connectionId string, addr net.IP) GSSAPIServer {
	backend := client.backend.GSSAPIConfig(connectionId, addr)
	if backend == nil {
		return nil
	}
	return &authzContext{
		client: client,
		backend: backend,
		connectionID: connectionId,
		remoteAddr: addr,
	}
}

func (client *httpAuthzClient) processAuthzWithRetry(
	princUsername string,
	loginUsername string,
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	authRequest := auth.AuthorizationRequest{
		PrincipalUsername: princUsername,
		LoginUsername: loginUsername,
		RemoteAddress: remoteAddr.String(),
		ConnectionID: connectionID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()
	var lastError error
	var lastLabels []metrics.MetricLabel
	logger := client.logger.
		WithLabel("connectionId", connectionID).
		WithLabel("princUsername", princUsername).
		WithLabel("loginUsername", loginUsername).
		WithLabel("url", client.endpoint)
loop:
	for {
		lastLabels = []metrics.MetricLabel{
			metrics.Label("authtype", "authorization"),
		}
		if lastError != nil {
			lastLabels = append(
				lastLabels,
				metrics.Label("retry", "1"),
			)
		} else {
			lastLabels = append(
				lastLabels,
				metrics.Label("retry", "0"),
			)
		}
		client.logAttempt(logger, lastLabels)

		authResponse := &auth.ResponseBody{}
		lastError = client.authServerRequest(client.endpoint, authRequest, authResponse)
		if lastError == nil {
			client.logAuthResponse(logger, authResponse, lastLabels, remoteAddr)
			return &httpAuthContext{authResponse.Success, authResponse.Metadata, nil}
		}
		reason := client.getReason(lastError)
		lastLabels = append(lastLabels, metrics.Label("reason", reason))
		client.logTemporaryFailure(logger, lastError, reason, lastLabels)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	return client.logAndReturnPermanentFailure(lastError, lastLabels, logger)
}

func (client *httpAuthzClient) logAttempt(logger log.Logger, lastLabels []metrics.MetricLabel) {
	logger.Debug(
		message.NewMessage(
			message.MAuth,
			"Authorization request",
		),
	)
	client.backendRequestsMetric.Increment(lastLabels...)
}

func (client *httpAuthzClient) logAndReturnPermanentFailure(
	lastError error,
	lastLabels []metrics.MetricLabel,
	logger log.Logger,
) AuthenticationContext {
	err := message.Wrap(
		lastError,
		message.EAuthBackendError,
		"Backend request for authorization failed, giving up",
	)
	client.backendFailureMetric.Increment(
		append(
			[]metrics.MetricLabel{
				metrics.Label("type", "hard"),
			}, lastLabels...,
		)...,
	)
	logger.Error(err)
	return &httpAuthContext{false, nil, err}
}

func (client *httpAuthzClient) logTemporaryFailure(
	logger log.Logger,
	lastError error,
	reason string,
	lastLabels []metrics.MetricLabel,
) {
	logger.Debug(
		message.Wrap(
			lastError,
			message.EAuthBackendError,
			"authorization request to backend failed, retrying in 10 seconds",
		).
			Label("reason", reason),
	)
	client.backendFailureMetric.Increment(
		append(
			[]metrics.MetricLabel{
				metrics.Label("type", "soft"),
			}, lastLabels...,
		)...,
	)
}

func (client *httpAuthzClient) getReason(lastError error) string {
	var typedErr message.Message
	reason := message.EUnknownError
	if errors.As(lastError, &typedErr) {
		reason = typedErr.Code()
	}
	return reason
}

func (client *httpAuthzClient) logAuthResponse(
	logger log.Logger,
	authResponse *auth.ResponseBody,
	labels []metrics.MetricLabel,
	remoteAddr net.IP,
) {
	if authResponse.Success {
		logger.Debug(
			message.NewMessage(
				message.MAuthSuccessful,
				"authorization successful",
			),
		)
		client.authSuccessMetric.Increment(remoteAddr, labels...)
	} else {
		logger.Debug(
			message.NewMessage(
				message.EAuthFailed,
				"authorization failed",
			),
		)
		client.authFailureMetric.Increment(remoteAddr, labels...)
	}
}

func (client *httpAuthzClient) authServerRequest(endpoint string, requestObject interface{}, response interface{}) error {
	statusCode, err := client.httpClient.Post(endpoint, requestObject, response)
	if err != nil {
		return err
	}
	if statusCode != 200 {
		return message.UserMessage(
			message.EAuthInvalidStatus,
			"Cannot authenticate at this time.",
			"auth server responded with an invalid status code: %d",
			statusCode,
		)
	}
	return nil
}
