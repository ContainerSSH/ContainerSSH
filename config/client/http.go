package client

import (
	"bytes"
	"containerssh/config"
	containerhttp "containerssh/http"
	"containerssh/protocol"
	"encoding/json"
	"net/http"
)

type HttpConfigClient struct {
	httpClient http.Client
	endpoint   string
}

func NewHttpConfigClient(config config.ConfigServerConfig) (ConfigClient, error) {
	if config.Url == "" {
		return nil, nil
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

	return &HttpConfigClient{
		httpClient: *realClient,
		endpoint:   config.Url,
	}, nil
}

func (client *HttpConfigClient) GetConfig(request protocol.ConfigRequest) (*protocol.ConfigResponse, error) {
	response := protocol.ConfigResponse{}
	err := client.configServerRequest(request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (client *HttpConfigClient) configServerRequest(requestObject interface{}, response interface{}) error {
	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(requestObject)
	if err != nil {
		//This is a bug
		return err
	}
	req, err := http.NewRequest(http.MethodGet, client.endpoint, buffer)
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
