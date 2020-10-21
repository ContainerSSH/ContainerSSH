package kuberun

import (
	backendMetrics "github.com/janoszen/containerssh/backend/metrics"
	"github.com/janoszen/containerssh/metrics"
)

var MetricBackendError = metrics.Metric{
	Name:   backendMetrics.MetricNameBackendError,
	Labels: map[string]string{backendMetrics.MetricLabelBackend: "kuberun"},
}
