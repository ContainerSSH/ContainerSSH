package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/janoszen/containerssh/config"
	containerhttp "github.com/janoszen/containerssh/http"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/metrics"
	"github.com/janoszen/containerssh/protocol"
	"net/http"
)

var MetricNameAuthBackendFailure = "auth_backend_failures"
var MetricAuthBackendFailure = metrics.Metric{
	Name:   MetricNameAuthBackendFailure,
	Labels: map[string]string{},
}
var MetricNameAuthFailure = "auth_failures"
var MetricAuthFailure = metrics.Metric{
	Name:   MetricNameAuthFailure,
	Labels: map[string]string{},
}
var MetricNameAuthSuccess = "auth_successes"
var MetricAuthSuccess = metrics.Metric{
	Name:   MetricNameAuthFailure,
	Labels: map[string]string{},
}

type HttpAuthClient struct {
	httpClient http.Client
	endpoint   string
	logger     log.Logger
	metric     *metrics.MetricCollector
}

func NewHttpAuthClient(
	config config.AuthConfig,
	logger log.Logger,
	metric *metrics.MetricCollector,
) (*HttpAuthClient, error) {
	if config.Url == "" {
		return nil, fmt.Errorf("no authentication server URL provided")
	}
	realClient, err := containerhttp.NewHttpClient(
		config.Timeout,
		config.CaCert,
		config.ClientCert,
		config.ClientKey,
		config.Url,
		logger,
	)
	if err != nil {
		return nil, err
	}

	metric.SetMetricMeta(MetricNameAuthBackendFailure, "Number of request failures to the authentication backend", metrics.MetricTypeCounter)
	metric.Set(MetricAuthBackendFailure, 0)
	metric.SetMetricMeta(MetricNameAuthFailure, "Number of failed authentications", metrics.MetricTypeCounter)
	metric.Set(MetricAuthFailure, 0)
	metric.SetMetricMeta(MetricNameAuthSuccess, "Number of successful authentications", metrics.MetricTypeCounter)
	metric.Set(MetricAuthSuccess, 0)

	return &HttpAuthClient{
		httpClient: *realClient,
		endpoint:   config.Url,
		logger:     logger,
		metric:     metric,
	}, nil
}

func (client *HttpAuthClient) Password(
	//Username provided
	username string,
	//Password provided
	password []byte,
	//Opaque session ID to identify the login attempt
	sessionId []byte,
	//Remote address in IP:port format
	remoteAddr string,
) (*protocol.AuthResponse, error) {
	client.logger.DebugF("Password authentication attempt user %s with public key for connection from %s", username, remoteAddr)
	authRequest := protocol.PasswordAuthRequest{
		User:          username,
		Username:      username,
		RemoteAddress: remoteAddr,
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		Password:      base64.StdEncoding.EncodeToString(password),
	}
	authResponse := &protocol.AuthResponse{}
	err := client.authServerRequest(client.endpoint+"/password", authRequest, authResponse)
	if err != nil {
		client.logger.DebugF("Failed password authentication for user %s with password for connection from %s", username, remoteAddr)
		return nil, err
	}
	client.logger.DebugF("Completed password authentication for user %s with password for connection from %s", username, remoteAddr)
	if authResponse.Success {
		client.logger.DebugF("Authentication successful %s with password for connection from %s", username, remoteAddr)
		client.metric.Increment(MetricAuthSuccess)
	} else {
		client.logger.DebugF("Authentication failed %s with password for connection from %s", username, remoteAddr)
		client.metric.Increment(MetricAuthFailure)
	}
	return authResponse, nil
}
func (client *HttpAuthClient) PubKey(
	//Username provided
	username string,
	//Serialized key data in SSH wire format
	pubKey []byte,
	//Opaque session ID to identify the login attempt
	sessionId []byte,
	//Remote address in IP:port format
	remoteAddr string,
) (*protocol.AuthResponse, error) {
	client.logger.DebugF("Public key authentication attempt user %s with public key for connection from %s", username, remoteAddr)
	authRequest := protocol.PublicKeyAuthRequest{
		User:          username,
		Username:      username,
		RemoteAddress: remoteAddr,
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		PublicKey:     base64.StdEncoding.EncodeToString(pubKey),
	}
	authResponse := &protocol.AuthResponse{}
	err := client.authServerRequest(client.endpoint+"/pubkey", authRequest, authResponse)
	if err != nil {
		client.logger.DebugF("Failed public key authentication for user %s with public key for connection from %s", username, remoteAddr)
		return nil, err
	}
	client.logger.DebugF("Completed password authentication for user %s with public key for connection from %s", username, remoteAddr)
	if authResponse.Success {
		client.logger.DebugF("Authentication successful %s with public key for connection from %s", username, remoteAddr)
		client.metric.Increment(MetricAuthSuccess)
	} else {
		client.logger.DebugF("Authentication failed %s with public key for connection from %s", username, remoteAddr)
		client.metric.Increment(MetricAuthFailure)
	}
	return authResponse, nil
}

func (client *HttpAuthClient) authServerRequest(endpoint string, requestObject interface{}, response interface{}) error {
	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(requestObject)
	if err != nil {
		//This is a bug
		return err
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, buffer)
	if err != nil {
		client.metric.Increment(MetricAuthBackendFailure)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		client.metric.Increment(MetricAuthBackendFailure)
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(response)
	if err != nil {
		client.metric.Increment(MetricAuthBackendFailure)
		return err
	}
	return nil
}
