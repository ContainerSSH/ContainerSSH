package metrics

type MetricType string

//goland:noinspection GoUnusedConst
const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)
