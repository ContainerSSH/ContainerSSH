package webhook

import (
	"net/http"

	"github.com/containerssh/containerssh/internal/config"
	"github.com/containerssh/containerssh/log"
)

// NewHandler creates a HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h ConfigRequestHandler, logger log.Logger) (http.Handler, error) {
	return config.NewHandler(h, logger)
}
