package util

import (
	"github.com/containerssh/containerssh/config"
	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
)

func Merge(config1 *config.AppConfig, config2 *config.AppConfig) (*config.AppConfig, error) {
	newConfig := &config.AppConfig{}
	err := copier.Copy(newConfig, config1)
	if err != nil {
		return nil, err
	}
	err = mergo.Merge(newConfig, config2)
	if err != nil {
		return nil, err
	}
	return newConfig, nil
}
