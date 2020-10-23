package util

import (
	"github.com/containerssh/containerssh/config"
	"github.com/creasty/defaults"
)

func GetDefaultConfig() (*config.AppConfig, error) {
	appConfig := &config.AppConfig{}
	err := defaults.Set(appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}
