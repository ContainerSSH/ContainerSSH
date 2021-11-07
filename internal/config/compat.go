package config

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

// fixCompatibility moves deprecated options to their new places and issues warnings.
func fixCompatibility(cfg *config.AppConfig, logger log.Logger) error {
	//goland:noinspection GoDeprecation
	if cfg.Auth.HTTPClientConfiguration.URL != "" {
		if cfg.Auth.Webhook.HTTPClientConfiguration.URL == "" {
			logger.Warning(
				message.NewMessage(
					message.WConfigAuthURLDeprecated,
					"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. Please use the new 'auth -> webhook -> url' option. See https://containerssh.io/deprecations/authurl for details.",
				))
			//goland:noinspection GoDeprecation
			cfg.Auth.Webhook.HTTPClientConfiguration = cfg.Auth.HTTPClientConfiguration
			//goland:noinspection GoDeprecation
			cfg.Auth.HTTPClientConfiguration = config.HTTPClientConfiguration{}
		} else {
			logger.Warning(
				message.NewMessage(
					message.WConfigAuthURLDeprecated,
					"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. The new option under 'auth -> webhook' takes precedence. See https://containerssh.io/deprecations/authurl for details.",
				))
		}
	}
	return nil
}
