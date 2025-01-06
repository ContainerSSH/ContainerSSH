package security

import (
	"fmt"

    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/sshserver"
    "go.containerssh.io/containerssh/log"
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
