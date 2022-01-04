package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

type httpAuthClient struct {
	timeout               time.Duration
	httpClient            http.Client
	endpoint              string
	logger                log.Logger
	metrics               metrics.Collector
	backendRequestsMetric metrics.SimpleCounter
	backendFailureMetric  metrics.SimpleCounter
	authSuccessMetric     metrics.GeoCounter
	authFailureMetric     metrics.GeoCounter
	enablePassword        bool
	enablePubKey          bool
}

type httpAuthContext struct {
	success  bool
	metadata *auth.ConnectionMetadata
	err      error
}

func (h httpAuthContext) Success() bool {
	return h.success
}

func (h httpAuthContext) Error() error {
	return h.err
}

func (h httpAuthContext) Metadata() *auth.ConnectionMetadata {
	return h.metadata
}

func (h httpAuthContext) OnDisconnect() {
}

func (client *httpAuthClient) KeyboardInteractive(
	_ string,
	_ func(instruction string, questions KeyboardInteractiveQuestions) (
		answers KeyboardInteractiveAnswers,
		err error,
	),
	_ string,
	_ net.IP,
) AuthenticationContext {
	return &httpAuthContext{false, nil, message.UserMessage(
		message.EAuthUnsupported,
		"Keyboard-interactive authentication is not available.",
		"Webhook authentication doesn't support keyboard-interactive.",
	)}
}

func (client *httpAuthClient) Password(
	username string,
	password []byte,
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	if !client.enablePassword {
		err := message.UserMessage(
			message.EAuthDisabled,
			"Password authentication failed.",
			"Password authentication is disabled.",
		)
		client.logger.Debug(err)
		return &httpAuthContext{false, nil, err}
	}
	url := client.endpoint + "/password"
	method := "Password"
	authType := "password"
	authRequest := auth.PasswordAuthRequest{
		Username:      username,
		RemoteAddress: remoteAddr.String(),
		ConnectionID:  connectionID,
		SessionID:     connectionID,
		Password:      base64.StdEncoding.EncodeToString(password),
	}

	return client.processAuthWithRetry(username, method, authType, connectionID, url, authRequest, remoteAddr)
}

func (client *httpAuthClient) PubKey(
	username string,
	pubKey string,
	connectionID string,
	remoteAddr net.IP,
) AuthenticationContext {
	if !client.enablePubKey {
		err := message.UserMessage(
			message.EAuthDisabled,
			"Public key authentication failed.",
			"Public key authentication is disabled.",
		)
		client.logger.Debug(err)
		return &httpAuthContext{false, nil, err}
	}
	url := client.endpoint + "/pubkey"
	authRequest := auth.PublicKeyAuthRequest{
		Username:      username,
		RemoteAddress: remoteAddr.String(),
		ConnectionID:  connectionID,
		SessionID:     connectionID,
		PublicKey:     pubKey,
	}
	method := "Public key"
	authType := "pubkey"

	return client.processAuthWithRetry(username, method, authType, connectionID, url, authRequest, remoteAddr)
}

func (client *httpAuthClient) GSSAPIConfig(connectionId string, addr net.IP) GSSAPIServer {
	return nil
}

func (client *httpAuthClient) processAuthWithRetry(
	username string,
	method string,
	authType string,
	connectionID string,
	url string,
	authRequest interface{},
	remoteAddr net.IP,
) AuthenticationContext {
	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()
	var lastError error
	var lastLabels []metrics.MetricLabel
	logger := client.logger.
		WithLabel("connectionId", connectionID).
		WithLabel("username", username).
		WithLabel("url", url).
		WithLabel("authtype", authType)
loop:
	for {
		lastLabels = []metrics.MetricLabel{
			metrics.Label("authtype", authType),
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
		client.logAttempt(logger, method, lastLabels)

		authResponse := &auth.ResponseBody{}
		lastError = client.authServerRequest(url, authRequest, authResponse)
		if lastError == nil {
			client.logAuthResponse(logger, method, authResponse, lastLabels, remoteAddr)
			return &httpAuthContext{authResponse.Success, authResponse.Metadata, nil}
		}
		reason := client.getReason(lastError)
		lastLabels = append(lastLabels, metrics.Label("reason", reason))
		client.logTemporaryFailure(logger, lastError, method, reason, lastLabels)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	return client.logAndReturnPermanentFailure(lastError, method, lastLabels, logger)
}

func (client *httpAuthClient) logAttempt(logger log.Logger, method string, lastLabels []metrics.MetricLabel) {
	logger.Debug(
		message.NewMessage(
			message.MAuth,
			"%s authentication request",
			method,
		),
	)
	client.backendRequestsMetric.Increment(lastLabels...)
}

func (client *httpAuthClient) logAndReturnPermanentFailure(
	lastError error,
	method string,
	lastLabels []metrics.MetricLabel,
	logger log.Logger,
) AuthenticationContext {
	err := message.Wrap(
		lastError,
		message.EAuthBackendError,
		"Backend request for %s authentication failed, giving up",
		strings.ToLower(method),
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

func (client *httpAuthClient) logTemporaryFailure(
	logger log.Logger,
	lastError error,
	method string,
	reason string,
	lastLabels []metrics.MetricLabel,
) {
	logger.Debug(
		message.Wrap(
			lastError,
			message.EAuthBackendError,
			"%s authentication request to backend failed, retrying in 10 seconds",
			method,
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

func (client *httpAuthClient) getReason(lastError error) string {
	var typedErr message.Message
	reason := message.EUnknownError
	if errors.As(lastError, &typedErr) {
		reason = typedErr.Code()
	}
	return reason
}

func (client *httpAuthClient) logAuthResponse(
	logger log.Logger,
	method string,
	authResponse *auth.ResponseBody,
	labels []metrics.MetricLabel,
	remoteAddr net.IP,
) {
	if authResponse.Success {
		logger.Debug(
			message.NewMessage(
				message.MAuthSuccessful,
				"%s authentication successful",
				method,
			),
		)
		client.authSuccessMetric.Increment(remoteAddr, labels...)
	} else {
		logger.Debug(
			message.NewMessage(
				message.EAuthFailed,
				"%s authentication failed",
				method,
			),
		)
		client.authFailureMetric.Increment(remoteAddr, labels...)
	}
}

func (client *httpAuthClient) authServerRequest(endpoint string, requestObject interface{}, response interface{}) error {
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
