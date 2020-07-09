package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/janoszen/containerssh/config"
	containerhttp "github.com/janoszen/containerssh/http"
	"github.com/janoszen/containerssh/protocol"
	"github.com/sirupsen/logrus"
	"net/http"
)

type HttpAuthClient struct {
	httpClient http.Client
	endpoint   string
}

func NewHttpAuthClient(config config.AuthConfig) (*HttpAuthClient, error) {
	if config.Url == "" {
		return nil, fmt.Errorf("no authentication server URL provided")
	}
	realClient, err := containerhttp.NewHttpClient(
		config.Timeout,
		config.CaCert,
		config.ClientCert,
		config.ClientKey,
		config.Url,
	)
	if err != nil {
		return nil, err
	}

	return &HttpAuthClient{
		httpClient: *realClient,
		endpoint:   config.Url,
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
	logrus.Tracef("Authentication user %s with password for connection from %s", username, remoteAddr)
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
		logrus.Tracef("Authentication failed (%s)", err)
		return nil, err
	}
	logrus.Tracef("Authentication successful")
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
	logrus.Tracef("Authentication user %s with public key for connection from %s", username, remoteAddr)
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
		logrus.Tracef("Authentication failed (%s)", err)
		return nil, err
	}
	logrus.Tracef("Authentication successful")
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
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(response)
	if err != nil {
		return err
	}
	return nil
}
