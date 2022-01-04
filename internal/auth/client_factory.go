package auth

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
)

func NewClient(
	cfg config.AuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (Client, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	var client Client
	var service service.Service
	var err error
	switch cfg.Method {
	case config.AuthMethodWebhook:
		client, err = NewHttpAuthClient(cfg, logger, metrics)
	case config.AuthMethodOAuth2:
		client, service, err = NewOAuth2Client(cfg, logger, metrics)
	case config.AuthMethodKerberos:
		client, err = NewKerberosClient(cfg, logger, metrics)
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
	if err != nil {
		return nil, nil, err
	}

	if cfg.Authz.Enable {
		client, err = NewHttpAuthzClient(client, cfg, logger, metrics)
	}
	return client, service, err
}
