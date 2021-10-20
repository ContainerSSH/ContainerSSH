package config

import (
	goHttp "net/http"

	"github.com/containerssh/containerssh/http"
	"github.com/containerssh/containerssh/log"
)

// NewHandler creates a HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h RequestHandler, logger log.Logger) (goHttp.Handler, error) {
	return http.NewServerHandler(&handler{
		handler: h,
		logger:  logger,
	}, logger), nil
}
