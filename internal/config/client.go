package config

import (
	"context"
	"net"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	// Get fetches the user-specific configuration.
	Get(
		ctx context.Context,
		username string,
		remoteAddr net.TCPAddr,
		connectionID string,
		metadata *auth.ConnectionMetadata,
	) (config.AppConfig, error)
}
