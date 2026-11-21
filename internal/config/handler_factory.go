package config

import (
	goHttp "net/http"

	"go.containerssh.io/containerssh/http"
	"go.containerssh.io/containerssh/log"
)

// NewHandler creates an HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h RequestHandler, logger log.Logger) (goHttp.Handler, error) {
	return http.NewServerHandler(&handler{
		handler: h,
		logger:  logger,
	}, logger), nil
}
