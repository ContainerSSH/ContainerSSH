package webhook

import (
    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/http"
    "go.containerssh.io/containerssh/internal/auth"
    "go.containerssh.io/containerssh/log"
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
