package config

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

// fixCompatibility moves deprecated options to their new places and issues warnings.
func fixCompatibility(cfg *config.AppConfig, logger log.Logger) error {
	if cfg.Listen != "" {
		if cfg.SSH.Listen == "" || cfg.SSH.Listen == "0.0.0.0:2222" {
			logger.Warning(
				message.NewMessage(
					message.WConfigListenDeprecated,
					"You are using the 'listen' option deprecated in ContainerSSH 0.4. Please use the new 'ssh -> listen' option. See https://containerssh.io/deprecations/listen for details.",
				))
			cfg.SSH.Listen = cfg.Listen
			cfg.Listen = ""
		} else {
			logger.Warning(
				message.NewMessage(
					message.WConfigListenDeprecated,
					"You are using the 'listen' option deprecated in ContainerSSH 0.4 as well as the new 'ssh -> listen' option. The new option takes precedence. Please see https://containerssh.io/deprecations/listen for details.",
				))
		}
	}
	return nil
}
