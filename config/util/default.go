package util

import (
	"github.com/creasty/defaults"
	"github.com/janoszen/containerssh/config"
)

func GetDefaultConfig() (*config.AppConfig, error) {
	appConfig := &config.AppConfig{}
	err := defaults.Set(appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}
