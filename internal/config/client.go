package config

import (
	"context"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/metadata"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	// Get fetches the user-specific configuration.
	Get(
		ctx context.Context,
		metadata metadata.ConnectionAuthenticatedMetadata,
	) (config.AppConfig, metadata.ConnectionAuthenticatedMetadata, error)
}
