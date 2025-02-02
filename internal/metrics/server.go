package metrics

import (
    "go.containerssh.io/containerssh/config"
    http2 "go.containerssh.io/containerssh/http"
    "go.containerssh.io/containerssh/log"
    messageCodes "go.containerssh.io/containerssh/message"
)

// NewServer creates a new metrics server based on the configuration. It MAY return nil if the server is disabled.
func NewServer(cfg config.MetricsConfig, collector Collector, logger log.Logger) (http2.Server, error) {
	if !cfg.Enable {
		return nil, nil
	}
	return http2.NewServer(
		"Metrics server",
		cfg.HTTPServerConfiguration,
		NewHandler(
			cfg.Path,
			collector,
		),
		logger,
		func(url string) {
			logger.Info(
				messageCodes.NewMessage(
					messageCodes.MHealthServiceAvailable,
					"Metrics server is now available at %s%s",
					url, cfg.Path,
				))
		},
	)
}
