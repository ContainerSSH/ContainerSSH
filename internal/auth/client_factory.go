package auth

import (
	"fmt"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/service"
)

func NewClient(
	cfg config.AuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (Client, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.AuthMethodWebhook:
		client, err := NewHttpAuthClient(cfg, logger, metrics)
		return client, nil, err
	case config.AuthMethodOAuth2:
		return NewOAuth2Client(cfg, logger, metrics)
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}
