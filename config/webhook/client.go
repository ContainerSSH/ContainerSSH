package webhook

import (
	"github.com/containerssh/libcontainerssh/internal/config"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	config.Client
}
