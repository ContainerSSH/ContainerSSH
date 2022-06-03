package config

import (
	"context"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/metadata"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	// Get fetches the user-specific configuration.
	Get(
		ctx context.Context,
		metadata metadata.ConnectionAuthenticatedMetadata,
	) (config.AppConfig, metadata.ConnectionAuthenticatedMetadata, error)
}
