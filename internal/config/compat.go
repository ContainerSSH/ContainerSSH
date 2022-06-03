package config

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
)

// fixCompatibility moves deprecated options to their new places and issues warnings.
func fixCompatibility(cfg *config.AppConfig, logger log.Logger) error {
	//goland:noinspection GoDeprecation
	if cfg.Auth.HTTPClientConfiguration.URL != "" {
		if cfg.Auth.PasswordAuth.Webhook.HTTPClientConfiguration.URL == "" {
			logger.Warning(
				message.NewMessage(
					message.WConfigAuthURLDeprecated,
					"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. Please use the new 'auth -> password -> webhook -> url' or `auth -> pubkey -> webhook -> url' options. See https://containerssh.io/deprecations/authurl for details.",
				))
			cfg.Auth.PasswordAuth.Method = config.PasswordAuthMethodWebhook
			cfg.Auth.PublicKeyAuth.Method = config.PubKeyAuthMethodWebhook
			//goland:noinspection GoDeprecation
			cfg.Auth.PasswordAuth.Webhook.HTTPClientConfiguration = cfg.Auth.HTTPClientConfiguration
			cfg.Auth.PublicKeyAuth.Webhook.HTTPClientConfiguration = cfg.Auth.HTTPClientConfiguration
			//goland:noinspection GoDeprecation
			cfg.Auth.HTTPClientConfiguration = config.HTTPClientConfiguration{}
		} else {
			logger.Warning(
				message.NewMessage(
					message.WConfigAuthURLDeprecated,
					"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. The new options under 'auth -> password -> webhook -> url' or `auth -> pubkey -> webhook -> url' take precedence. See https://containerssh.io/deprecations/authurl for details.",
				))
		}
	}
	return nil
}
