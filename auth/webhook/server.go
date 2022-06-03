package webhook

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/internal/auth"
    "go.containerssh.io/libcontainerssh/log"
)

// NewServer returns a complete HTTP server that responds to the authentication requests.
func NewServer(
	cfg config.HTTPServerConfiguration,
	h AuthRequestHandler,
	logger log.Logger,
) (http.Server, error) {
	return auth.NewServer(
		cfg,
		h,
		logger,
	)
}
