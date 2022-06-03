package webhook

import (
	"net/http"

    "go.containerssh.io/libcontainerssh/internal/auth"
    "go.containerssh.io/libcontainerssh/log"
)

// NewHandler creates a HTTP handler that forwards calls to the provided h config request handler.
func NewHandler(h AuthRequestHandler, logger log.Logger) http.Handler {
	return auth.NewHandler(h, logger)
}
