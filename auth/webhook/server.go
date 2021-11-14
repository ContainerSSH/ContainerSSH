package webhook

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/log"
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
