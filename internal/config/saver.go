package config

import (
	"go.containerssh.io/libcontainerssh/config"
)

// ConfigSaver is a utility to store configuration
type ConfigSaver interface {
	// Save stores the passed configuration and returns an error on failure.
	Save(config *config.AppConfig) error
}
