package config

import "github.com/containerssh/containerssh/log"

// swagger:model
type AppLogConfig struct {
	// Minimum log level to log. Everything below the specified level will not be logged. Can be changed from the configuration server.
	Level log.LevelString `json:"level" yaml:"level" default:"info" comment:"Log level. Can be any valid Syslog log level."`
}
