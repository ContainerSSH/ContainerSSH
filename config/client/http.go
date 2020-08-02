package client

import (
	"bytes"
	"encoding/json"
	"github.com/janoszen/containerssh/config"
	containerhttp "github.com/janoszen/containerssh/http"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/protocol"
	"net/http"
)

type HttpConfigClient struct {
	httpClient http.Client
	endpoint   string
	logger     log.Logger
}

func NewHttpConfigClient(
	config config.ConfigServerConfig,
	logger log.Logger,
) (ConfigClient, error) {
	if config.Url == "" {
		return nil, nil
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

	return &HttpConfigClient{
		httpClient: *realClient,
		endpoint:   config.Url,
		logger: logger,
	}, nil
}

func (client *HttpConfigClient) GetConfig(request protocol.ConfigRequest) (*protocol.ConfigResponse, error) {
	client.logger.DebugF("Fetching configuration for connection for user %s", request.Username)
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
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(response)
	if err != nil {
		return err
	}
	return nil
}
