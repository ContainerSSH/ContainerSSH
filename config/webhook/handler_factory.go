package webhook

import (
	"net/http"

    "go.containerssh.io/libcontainerssh/internal/config"
    "go.containerssh.io/libcontainerssh/log"
)

// NewHandler creates a HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h ConfigRequestHandler, logger log.Logger) (http.Handler, error) {
	return config.NewHandler(h, logger)
}
