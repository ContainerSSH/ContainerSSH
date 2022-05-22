package config

import (
	"context"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/metadata"
)

// Loader is a utility to load and update an existing configuration structure.
type Loader interface {
	// Load loads the configuration from a generic configuration source.
	//
	// - ctx is the deadline for loading the configuration.
	// - config is the configuration structure to be loaded into.
	Load(
		ctx context.Context,
		config *config.AppConfig,
	) error

	// LoadConnection loads the configuration for a specific connection source.
	//
	// - ctx is the deadline for loading the configuration.
	// - meta is the metadata for the currenct connection
	// - config is the configuration struct to be loaded into.
	LoadConnection(
		ctx context.Context,
		meta metadata.ConnectionAuthenticatedMetadata,
		config *config.AppConfig,
	) (metadata.ConnectionAuthenticatedMetadata, error)
}
