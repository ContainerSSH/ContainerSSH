package kubernetes

import (
	"context"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
)

// kubernetesClientFactory creates a kubernetesClient based on a configuration
type kubernetesClientFactory interface {
	// get takes a configuration and returns a kubernetes client if the configuration was populated.
	// Returns an error if the configuration is invalid.
	get(ctx context.Context, config config.KubernetesConfig, logger log.Logger) (kubernetesClient, error)
}
