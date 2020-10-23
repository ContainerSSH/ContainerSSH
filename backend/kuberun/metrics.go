package kuberun

import (
	backendMetrics "github.com/containerssh/containerssh/backend/metrics"
	"github.com/containerssh/containerssh/metrics"
)

var MetricBackendError = metrics.Metric{
	Name:   backendMetrics.MetricNameBackendError,
	Labels: map[string]string{backendMetrics.MetricLabelBackend: "kuberun"},
}
