package auth

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

// NewHttpAuthClient creates a new HTTP authentication client
//goland:noinspection GoUnusedExportedFunction
func NewHttpAuthClient(
	cfg config.AuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (Client, error) {
	if cfg.Method != config.AuthMethodWebhook {
		return nil, fmt.Errorf("authentication is not set to webhook")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if cfg.URL != "" {
		logger.Warning(
			message.NewMessage(
				message.EAuthDeprecated,
				"The auth.url setting is deprecated, please switch to using auth.webhook.url. See https://containerssh.io/deprecations/authurl for details.",
			))
		//goland:noinspection GoDeprecation
		cfg.Webhook.HTTPClientConfiguration = cfg.HTTPClientConfiguration
		//goland:noinspection GoDeprecation
		cfg.Webhook.Password = cfg.Password
		//goland:noinspection GoDeprecation
		cfg.Webhook.PubKey = cfg.PubKey
		//goland:noinspection GoDeprecation
		cfg.HTTPClientConfiguration = config.HTTPClientConfiguration{}
	}

	realClient, err := http.NewClient(
		cfg.Webhook.HTTPClientConfiguration,
		logger,
	)
	if err != nil {
		return nil, err
	}

	backendRequestsMetric, backendFailureMetric, authSuccessMetric, authFailureMetric := createMetrics(metrics)
	return &httpAuthClient{
		enablePassword:        cfg.Webhook.Password,
		enablePubKey:          cfg.Webhook.PubKey,
		timeout:               cfg.AuthTimeout,
		httpClient:            realClient,
		logger:                logger,
		metrics:               metrics,
		backendRequestsMetric: backendRequestsMetric,
		backendFailureMetric:  backendFailureMetric,
		authSuccessMetric:     authSuccessMetric,
		authFailureMetric:     authFailureMetric,
	}, nil
}

func createMetrics(metrics metrics.Collector) (
	metrics.Counter,
	metrics.Counter,
	metrics.GeoCounter,
	metrics.GeoCounter,
) {
	backendRequestsMetric := metrics.MustCreateCounter(
		MetricNameAuthBackendRequests,
		"requests_total",
		"The number of requests sent to the configuration server.",
	)
	backendFailureMetric := metrics.MustCreateCounter(
		MetricNameAuthBackendFailure,
		"failures_total",
		"The number of request failures to the configuration server.",
	)
	authSuccessMetric := metrics.MustCreateCounterGeo(
		MetricNameAuthSuccess,
		"success_total",
		"The number of successful authentications.",
	)
	authFailureMetric := metrics.MustCreateCounterGeo(
		MetricNameAuthFailure,
		"failures_total",
		"The number of failed authentications.",
	)
	return backendRequestsMetric, backendFailureMetric, authSuccessMetric, authFailureMetric
}
