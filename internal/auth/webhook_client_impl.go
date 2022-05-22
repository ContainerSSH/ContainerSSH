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
	"github.com/containerssh/libcontainerssh/metadata"
)

type webhookClient struct {
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
	enableAuthz           bool
}

func (client *webhookClient) Authorize(
	meta metadata.ConnectionAuthenticatedMetadata,
) AuthorizationResponse {
	if !client.enableAuthz {
		err := message.UserMessage(
			message.EAuthDisabled,
			"Authorization failed.",
			"Authorization is disabled.",
		)
		client.logger.Debug(err)
		return &webhookClientContext{meta.AuthFailed(), false, err}
	}
	return client.processAuthzWithRetry(meta)
}

func (client *webhookClient) Password(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) AuthenticationContext {
	if !client.enablePassword {
		err := message.UserMessage(
			message.EAuthDisabled,
			"Password authentication failed.",
			"Password authentication is disabled.",
		)
		client.logger.Debug(err)
		return &webhookClientContext{meta.AuthFailed(), false, err}
	}
	url := client.endpoint + "/password"
	method := "Password"
	authType := "password"
	authRequest := auth.PasswordAuthRequest{
		ConnectionAuthPendingMetadata: meta,
		Password:                      base64.StdEncoding.EncodeToString(password),
	}

	return client.processAuthWithRetry(meta, method, authType, url, authRequest)
}

func (client *webhookClient) PubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	pubKey auth.PublicKey,
) AuthenticationContext {
	if !client.enablePubKey {
		err := message.UserMessage(
			message.EAuthDisabled,
			"Public key authentication failed.",
			"Public key authentication is disabled.",
		)
		client.logger.Debug(err)
		return &webhookClientContext{meta.AuthFailed(), false, err}
	}
	url := client.endpoint + "/pubkey"
	authRequest := auth.PublicKeyAuthRequest{
		ConnectionAuthPendingMetadata: meta,
		PublicKey:                     pubKey,
	}
	method := "Public key"
	authType := "pubkey"

	return client.processAuthWithRetry(meta, method, authType, url, authRequest)
}

func (client *webhookClient) processAuthWithRetry(
	meta metadata.ConnectionAuthPendingMetadata,
	method string,
	authType string,
	url string,
	authRequest interface{},
) AuthenticationContext {
	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()
	var lastError error
	var lastLabels []metrics.MetricLabel
	logger := client.logger.
		WithLabel("connectionId", meta.ConnectionID).
		WithLabel("username", meta.Username).
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
			authenticatedMeta := meta.Authenticated("")
			authenticatedMeta.Merge(authResponse.ConnectionAuthenticatedMetadata)
			client.logAuthResponse(logger, method, authResponse, lastLabels, authenticatedMeta.RemoteAddress.IP)

			return &webhookClientContext{
				authenticatedMeta,
				authResponse.Success,
				nil,
			}
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
	return client.logAndReturnPermanentFailure(meta, lastError, method, lastLabels, logger)
}

func (client *webhookClient) logAttempt(logger log.Logger, method string, lastLabels []metrics.MetricLabel) {
	logger.Debug(
		message.NewMessage(
			message.MAuth,
			"%s authentication request",
			method,
		),
	)
	client.backendRequestsMetric.Increment(lastLabels...)
}

func (client *webhookClient) logAndReturnPermanentFailure(
	meta metadata.ConnectionAuthPendingMetadata,
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
	return &webhookClientContext{meta.AuthFailed(), false, err}
}

func (client *webhookClient) logTemporaryFailure(
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

func (client *webhookClient) getReason(lastError error) string {
	var typedErr message.Message
	reason := message.EUnknownError
	if errors.As(lastError, &typedErr) {
		reason = typedErr.Code()
	}
	return reason
}

func (client *webhookClient) logAuthResponse(
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

func (client *webhookClient) authServerRequest(endpoint string, requestObject interface{}, response interface{}) error {
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

func (client *webhookClient) processAuthzWithRetry(
	meta metadata.ConnectionAuthenticatedMetadata,
) AuthenticationContext {
	url := client.endpoint + "/authz"
	authzRequest := auth.AuthorizationRequest{}

	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()
	var lastError error
	var lastLabels []metrics.MetricLabel
	logger := client.logger.
		WithLabel("connectionId", meta.ConnectionID).
		WithLabel("authenticatedUsername", meta.AuthenticatedUsername).
		WithLabel("providedUsername", meta.Username).
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
		client.logAuthzAttempt(logger, lastLabels)

		authResponse := &auth.ResponseBody{}
		lastError = client.authServerRequest(url, authzRequest, authResponse)
		if lastError == nil {
			authenticatedMeta := meta.Authenticated("")
			authenticatedMeta.Merge(authResponse.ConnectionAuthenticatedMetadata)
			client.logAuthzResponse(authenticatedMeta, logger, authResponse, lastLabels)
			return &webhookClientContext{
				authenticatedMeta,
				authResponse.Success,
				nil,
			}
		}
		reason := client.getReason(lastError)
		lastLabels = append(lastLabels, metrics.Label("reason", reason))
		client.logTemporaryAuthzFailure(logger, lastError, reason, lastLabels)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	return client.logAndReturnPermanentAuthzFailure(meta, lastError, lastLabels, logger)
}

func (client *webhookClient) logAuthzAttempt(logger log.Logger, lastLabels []metrics.MetricLabel) {
	logger.Debug(
		message.NewMessage(
			message.MAuth,
			"Authorization request",
		),
	)
	client.backendRequestsMetric.Increment(lastLabels...)
}

func (client *webhookClient) logAuthzResponse(
	meta metadata.ConnectionAuthenticatedMetadata,
	logger log.Logger,
	authResponse *auth.ResponseBody,
	labels []metrics.MetricLabel,
) {
	if authResponse.Success {
		logger.Debug(
			message.NewMessage(
				message.MAuthSuccessful,
				"authorization successful",
			),
		)
		client.authSuccessMetric.Increment(meta.RemoteAddress.IP, labels...)
	} else {
		logger.Debug(
			message.NewMessage(
				message.EAuthFailed,
				"authorization failed",
			),
		)
		client.authFailureMetric.Increment(meta.RemoteAddress.IP, labels...)
	}
}

func (client *webhookClient) logAndReturnPermanentAuthzFailure(
	meta metadata.ConnectionAuthenticatedMetadata,
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
	return &webhookClientContext{meta.AuthFailed(), false, err}
}

func (client *webhookClient) logTemporaryAuthzFailure(
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
