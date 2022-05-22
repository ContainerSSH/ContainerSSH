package auth

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
)

// NewWebhookClient creates a new HTTP authentication client.
//goland:noinspection GoUnusedExportedFunction
func NewWebhookClient(
	authType AuthenticationType,
	cfg config.AuthWebhookClientConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (WebhookClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	realClient, err := http.NewClient(
		cfg.HTTPClientConfiguration,
		logger,
	)
	if err != nil {
		return nil, err
	}

	backendRequestsMetric, backendFailureMetric, authSuccessMetric, authFailureMetric := createMetrics(metrics)
	return &webhookClient{
		timeout:               cfg.AuthTimeout,
		httpClient:            realClient,
		logger:                logger,
		metrics:               metrics,
		backendRequestsMetric: backendRequestsMetric,
		backendFailureMetric:  backendFailureMetric,
		authSuccessMetric:     authSuccessMetric,
		authFailureMetric:     authFailureMetric,
		enablePassword:        authType == AuthenticationTypePassword || authType == AuthenticationTypeAll,
		enablePubKey:          authType == AuthenticationTypePublicKey || authType == AuthenticationTypeAll,
		enableAuthz:           authType == AuthenticationTypeAuthz || authType == AuthenticationTypeAll,
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
