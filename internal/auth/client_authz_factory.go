package auth

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
)

// NewHttpAuthClient creates a new HTTP authentication client
//goland:noinspection GoUnusedExportedFunction
func NewHttpAuthzClient(
	backend Client,
	cfg config.AuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (Client, error) {
	if !cfg.Authz.Enable{
		return nil, fmt.Errorf("authorization is disabled")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	httpClient, err := http.NewClient(
		cfg.Authz.HTTPClientConfiguration,
		logger,
	)
	if err != nil {
		return nil, err
	}

	backendRequestsMetric, backendFailureMetric, authSuccessMetric, authFailureMetric := createAuthzMetrics(metrics)
	return &httpAuthzClient{
		backend:               backend,
		timeout:               cfg.AuthTimeout,
		httpClient:            httpClient,
		logger:                logger,
		metrics:               metrics,
		backendRequestsMetric: backendRequestsMetric,
		backendFailureMetric:  backendFailureMetric,
		authSuccessMetric:     authSuccessMetric,
		authFailureMetric:     authFailureMetric,
	}, nil
}

func createAuthzMetrics(metrics metrics.Collector) (
	metrics.Counter,
	metrics.Counter,
	metrics.GeoCounter,
	metrics.GeoCounter,
) {
	backendRequestsMetric := metrics.MustCreateCounter(
		MetricNameAuthBackendRequests,
		"requests",
		"The number of requests sent to the configuration server.",
	)
	backendFailureMetric := metrics.MustCreateCounter(
		MetricNameAuthBackendFailure,
		"requests",
		"The number of request failures to the configuration server.",
	)
	authSuccessMetric := metrics.MustCreateCounterGeo(
		MetricNameAuthSuccess,
		"requests",
		"The number of successful authorizations.",
	)
	authFailureMetric := metrics.MustCreateCounterGeo(
		MetricNameAuthFailure,
		"requests",
		"The number of failed authorizations.",
	)
	return backendRequestsMetric, backendFailureMetric, authSuccessMetric, authFailureMetric
}
