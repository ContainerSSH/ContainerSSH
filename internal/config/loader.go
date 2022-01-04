package config

import (
	"context"
	"net"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
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
	// - username is the username from the SSH connection.
	// - remoteAddr is the source IP address and port of the connection.
	// - connectionID is an opaque ID made of hexadecimal numbers identifying the connection.
	// - config is the configuration struct to be loaded into.
	LoadConnection(
		ctx context.Context,
		username string,
		remoteAddr net.TCPAddr,
		connectionID string,
		metadata *auth.ConnectionMetadata,
		config *config.AppConfig,
	) error
}
