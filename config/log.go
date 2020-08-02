package config

import "github.com/janoszen/containerssh/log"

type LogConfig struct {
	Level log.LevelString `json:"level" yaml:"level" default:"info" comment:"Log level. Can be any valid Syslog log level."`
}
