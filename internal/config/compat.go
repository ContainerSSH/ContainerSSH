package config

import (
    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/message"
)

// fixCompatibility moves deprecated options to their new places and issues warnings.
func fixCompatibility(cfg *config.AppConfig, logger log.Logger) error {
	//goland:noinspection GoDeprecation
	if cfg.Auth.HTTPClientConfiguration.URL == "" {
		return nil
	}

	if cfg.Auth.PasswordAuth.Webhook.HTTPClientConfiguration.URL == "" {
		logger.Warning(
			message.NewMessage(
				message.WConfigAuthURLDeprecated,
				"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. Please use the new 'auth -> password -> webhook -> url' option. See https://containerssh.io/deprecations/authurl for details.",
			))
		//goland:noinspection GoDeprecation
		if cfg.Auth.Password != nil && *cfg.Auth.Password {
			cfg.Auth.PasswordAuth.Method = config.PasswordAuthMethodWebhook
			//goland:noinspection GoDeprecation
			cfg.Auth.PasswordAuth.Webhook.HTTPClientConfiguration = cfg.Auth.HTTPClientConfiguration
		}
	} else {
		logger.Warning(
			message.NewMessage(
				message.WConfigAuthURLDeprecated,
				"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. The new option under 'auth -> password -> webhook -> url' take precedence. See https://containerssh.io/deprecations/authurl for details.",
			))
	}

	if cfg.Auth.PublicKeyAuth.Webhook.HTTPClientConfiguration.URL == "" {
		logger.Warning(
			message.NewMessage(
				message.WConfigAuthURLDeprecated,
				"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. Please use the new 'auth -> pubkey -> webhook -> url' option. See https://containerssh.io/deprecations/authurl for details.",
			))
		//goland:noinspection GoDeprecation
		if cfg.Auth.PubKey != nil && *cfg.Auth.PubKey {
			cfg.Auth.PublicKeyAuth.Method = config.PubKeyAuthMethodWebhook
			//goland:noinspection GoDeprecation
			cfg.Auth.PublicKeyAuth.Webhook.HTTPClientConfiguration = cfg.Auth.HTTPClientConfiguration
		}
	} else {
		logger.Warning(
			message.NewMessage(
				message.WConfigAuthURLDeprecated,
				"You are using the 'auth.url' option deprecated in ContainerSSH 0.5. The new option under 'auth -> pubkey -> webhook -> url' take precedence. See https://containerssh.io/deprecations/authurl for details.",
			))
	}
	return nil
}
