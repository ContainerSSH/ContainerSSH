package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

type HttpClient struct {
	httpClient http.Client
	endpoint   string
}

func NewHttpClient(endpoint string) *HttpClient {
	return &HttpClient{
		httpClient: http.Client{
			Timeout: time.Second * 2,
		},
		endpoint: endpoint,
	}
}

func (client *HttpClient) Password(
	//Username provided
	username string,
	//Password provided
	password []byte,
	//Opaque session ID to identify the login attempt
	sessionId []byte,
	//Remote address in IP:port format
	remoteAddr string,
) (*Response, error) {
	authRequest := PasswordRequest{
		User:          username,
		RemoteAddress: remoteAddr,
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		Password:      base64.StdEncoding.EncodeToString(password),
	}
	authResponse := &Response{}
	err := client.authServerRequest(client.endpoint+"/password", authRequest, authResponse)
	if err != nil {
		return nil, err
	}
	return authResponse, nil
}
func (client *HttpClient) PubKey(
	//Username provided
	username string,
	//Serialized key data in SSH wire format
	pubKey []byte,
	//Opaque session ID to identify the login attempt
	sessionId []byte,
	//Remote address in IP:port format
	remoteAddr string,
) (*Response, error) {
	authRequest := PublicKeyRequest{
		User:          username,
		RemoteAddress: remoteAddr,
		SessionId:     base64.StdEncoding.EncodeToString(sessionId),
		PublicKey:     base64.StdEncoding.EncodeToString(pubKey),
	}
	authResponse := &Response{}
	err := client.authServerRequest(client.endpoint+"/pubkey", authRequest, authResponse)
	if err != nil {
		return nil, err
	}
	return authResponse, nil
}

func (client *HttpClient) authServerRequest(endpoint string, requestObject interface{}, response interface{}) error {
	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(requestObject)
	req, err := http.NewRequest(http.MethodGet, endpoint, buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}
	return nil
}
