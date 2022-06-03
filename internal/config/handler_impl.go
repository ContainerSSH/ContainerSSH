package config

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/log"
)

type handler struct {
	handler RequestHandler
	logger  log.Logger
}

func (h *handler) OnRequest(request http.ServerRequest, response http.ServerResponse) error {
	requestObject := config.Request{}
	if err := request.Decode(&requestObject); err != nil {
		return err
	}
	appConfig, err := h.handler.OnConfig(requestObject)
	if err != nil {
		return err
	}
	responseObject := config.ResponseBody{
		Config: appConfig,
	}
	response.SetBody(responseObject)
	response.SetStatus(200)
	return nil
}
