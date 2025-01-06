package webhook

import (
	"go.containerssh.io/containerssh/internal/config"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	config.Client
}
