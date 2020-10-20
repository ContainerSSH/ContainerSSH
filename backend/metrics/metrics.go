package metrics

import "github.com/janoszen/containerssh/metrics"

func Init(metric *metrics.MetricCollector) {
	metric.SetMetricMeta(MetricNameBackendError, "Number of errors in the backend", metrics.MetricTypeCounter)
}

var MetricLabelBackend = "backend"
var MetricNameBackendError = "backend_errors"
