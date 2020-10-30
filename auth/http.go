package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/containerssh/containerssh/config"
	containerhttp "github.com/containerssh/containerssh/http"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/metrics"
	"github.com/containerssh/containerssh/protocol"
	"net"
	"net/http"
)

var MetricNameAuthBackendFailure = "containerssh_auth_server_failures"
var MetricAuthBackendFailure = metrics.Metric{
	Name:   MetricNameAuthBackendFailure,
	Labels: map[string]string{},
}
var MetricNameAuthFailure = "containerssh_auth_failures"
var MetricAuthFailurePubkey = metrics.Metric{
	Name:   MetricNameAuthFailure,
	Labels: map[string]string{"authtype": "pubkey"},
}
var MetricAuthFailurePassword = metrics.Metric{
	Name:   MetricNameAuthFailure,
	Labels: map[string]string{"authtype": "password"},
}
var MetricNameAuthSuccess = "containerssh_auth_success"
var MetricAuthSuccessPubkey = metrics.Metric{
	Name:   MetricNameAuthSuccess,
	Labels: map[string]string{"authtype": "pubkey"},
}
var MetricAuthSuccessPassword = metrics.Metric{
	Name:   MetricNameAuthSuccess,
	Labels: map[string]string{"authtype": "password"},
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
	metric.SetMetricMeta(MetricNameAuthSuccess, "Number of successful authentications", metrics.MetricTypeCounter)

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
	remoteAddr net.IP,
) (*protocol.AuthResponse, error) {
	client.logger.DebugF("password authentication attempt user %s with public key for connection from %s", username, remoteAddr)
	authRequest := protocol.PasswordAuthRequest{
		User:          username,
		Username:      username,
		RemoteAddress: remoteAddr.String(),
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		Password:      base64.StdEncoding.EncodeToString(password),
	}
	authResponse := &protocol.AuthResponse{}
	err := client.authServerRequest(client.endpoint+"/password", authRequest, authResponse)
	if err != nil {
		client.logger.DebugF("failed password authentication for user %s with password for connection from %s", username, remoteAddr.String())
		return nil, err
	}
	client.logger.DebugF("completed password authentication for user %s with password for connection from %s", username, remoteAddr.String())
	if authResponse.Success {
		client.logger.DebugF("authentication successful %s with password for connection from %s", username, remoteAddr.String())
		client.metric.IncrementGeo(MetricAuthSuccessPassword, remoteAddr)
	} else {
		client.logger.DebugF("authentication failed %s with password for connection from %s", username, remoteAddr.String())
		client.metric.IncrementGeo(MetricAuthFailurePassword, remoteAddr)
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
	remoteAddr net.IP,
) (*protocol.AuthResponse, error) {
	client.logger.DebugF("public key authentication attempt user %s with public key for connection from %s", username, remoteAddr.String())
	authRequest := protocol.PublicKeyAuthRequest{
		User:          username,
		Username:      username,
		RemoteAddress: remoteAddr.String(),
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		PublicKey:     base64.StdEncoding.EncodeToString(pubKey),
	}
	authResponse := &protocol.AuthResponse{}
	err := client.authServerRequest(client.endpoint+"/pubkey", authRequest, authResponse)
	if err != nil {
		client.logger.DebugF("failed public key authentication for user %s with public key for connection from %s", username, remoteAddr.String())
		return nil, err
	}
	client.logger.DebugF("completed password authentication for user %s with public key for connection from %s", username, remoteAddr.String())
	if authResponse.Success {
		client.logger.DebugF("authentication successful %s with public key for connection from %s", username, remoteAddr.String())
		client.metric.IncrementGeo(MetricAuthSuccessPubkey, remoteAddr)
	} else {
		client.logger.DebugF("authentication failed %s with public key for connection from %s", username, remoteAddr.String())
		client.metric.IncrementGeo(MetricAuthFailurePubkey, remoteAddr)
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
