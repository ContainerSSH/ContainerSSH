package config

import (
	goHttp "net/http"

    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/log"
)

// NewHandler creates a HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h RequestHandler, logger log.Logger) (goHttp.Handler, error) {
	return http.NewServerHandler(&handler{
		handler: h,
		logger:  logger,
	}, logger), nil
}
