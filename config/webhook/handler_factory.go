package webhook

import (
	"net/http"

	"go.containerssh.io/containerssh/internal/config"
	"go.containerssh.io/containerssh/log"
)

// NewHandler creates an HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h ConfigRequestHandler, logger log.Logger) (http.Handler, error) {
	return config.NewHandler(h, logger)
}
