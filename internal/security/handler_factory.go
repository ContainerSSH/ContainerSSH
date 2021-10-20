package security

import (
	"fmt"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/log"
)

// New creates a new security backend proxy.
//goland:noinspection GoUnusedExportedFunction
func New(
	config config.SecurityConfig,
	backend sshserver.NetworkConnectionHandler,
	logger log.Logger,
) (sshserver.NetworkConnectionHandler, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid security configuration (%w)", err)
	}
	return &networkHandler{
		config:  config,
		backend: backend,
		logger:  logger,
	}, nil
}
