package config

import "github.com/janoszen/containerssh/log"

// swagger:model
type AppLogConfig struct {
	Level log.LevelString `json:"level" yaml:"level" default:"info" comment:"Log level. Can be any valid Syslog log level."`
}
