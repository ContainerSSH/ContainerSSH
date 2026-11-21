package config

import (
	"context"

    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/metrics"
    "go.containerssh.io/containerssh/internal/structutils"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/metadata"
)

// NewHTTPLoader loads configuration from HTTP servers for specific connections.
//goland:noinspection GoUnusedExportedFunction
func NewHTTPLoader(
	config config.ClientConfig,
	logger log.Logger,
	metricsCollector metrics.Collector,
) (Loader, error) {
	client, err := NewClient(config, logger, metricsCollector)
	if err != nil {
		return nil, err
	}
	return &httpLoader{
		client: client,
	}, nil
}

type httpLoader struct {
	client Client
}

func (h *httpLoader) Load(_ context.Context, _ *config.AppConfig) error {
	return nil
}

func (h *httpLoader) LoadConnection(
	ctx context.Context,
	meta metadata.ConnectionAuthenticatedMetadata,
	config *config.AppConfig,
) (metadata.ConnectionAuthenticatedMetadata, error) {
	newAppConfig, newMeta, err := h.client.Get(ctx, meta)
	if err != nil {
		return meta, err
	}
	if err := structutils.Merge(config, &newAppConfig); err != nil {
		return meta, err
	}
	return newMeta, err
}
