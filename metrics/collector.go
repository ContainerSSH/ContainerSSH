package metrics

import (
	"sort"
	"strings"
	"sync"
)

type MetricType string

//goland:noinspection GoUnusedConst
const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

type Metric struct {
	Name   string
	Labels map[string]string
}

func (metric *Metric) ToString() string {
	var labelList []string

	keys := make([]string, 0, len(metric.Labels))
	for k := range metric.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		labelList = append(labelList, k+"=\""+metric.Labels[k]+"\"")
	}

	var labels string
	if len(labelList) > 0 {
		labels = "{" + strings.Join(labelList, ",") + "}"
	} else {
		labels = ""
	}

	return metric.Name + labels
}

type MetricCollector struct {
	mutex      *sync.Mutex
	metricKeys map[string]*Metric
	metrics    map[string]map[*Metric]float64
	help       map[string]string
	types      map[string]MetricType
}

func New() *MetricCollector {
	return &MetricCollector{
		&sync.Mutex{},
		make(map[string]*Metric, 0),
		make(map[string]map[*Metric]float64, 0),
		make(map[string]string, 0),
		make(map[string]MetricType, 0),
	}
}

func (collector *MetricCollector) GetHelp(metricName string) string {
	if val, ok := collector.help[metricName]; ok {
		return val
	}
	return ""
}

func (collector *MetricCollector) GetType(metricName string) MetricType {
	if val, ok := collector.types[metricName]; ok {
		return val
	}
	return ""
}

func (collector *MetricCollector) GetMetricNames() []string {
	names := make([]string, len(collector.metrics))
	i := 0
	for k := range collector.metrics {
		names[i] = k
		i = i + 1
	}
	return names
}

func (collector *MetricCollector) GetMetrics(name string) map[*Metric]float64 {
	if val, ok := collector.metrics[name]; ok {
		return val
	} else {
		return make(map[*Metric]float64, 0)
	}
}

func (collector *MetricCollector) SetMetricMeta(metricName string, help string, metricType MetricType) {
	collector.mutex.Lock()
	collector.help[metricName] = help
	collector.types[metricName] = metricType
	collector.mutex.Unlock()
}

func (collector *MetricCollector) Get(metric Metric) float64 {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	return collector.get(metric)
}
func (collector *MetricCollector) Set(metric Metric, value float64) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	collector.set(metric, value)
}
func (collector *MetricCollector) Increment(metric Metric) float64 {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	value := collector.get(metric)
	value = value + 1
	collector.set(metric, value)
	return value
}

func (collector *MetricCollector) Decrement(metric Metric) float64 {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	value := collector.get(metric)
	value = value - 1
	collector.set(metric, value)
	return value
}

func (collector *MetricCollector) get(metric Metric) float64 {
	var metricKey *Metric
	if val, ok := collector.metricKeys[(&metric).ToString()]; ok {
		metricKey = val
	} else {
		collector.metricKeys[(&metric).ToString()] = &metric
		metricKey = collector.metricKeys[(&metric).ToString()]
	}
	if _, ok := collector.metrics[metric.Name]; !ok {
		collector.metrics[metric.Name] = make(map[*Metric]float64, 0)
		return 0
	}
	if _, ok := collector.metrics[metric.Name][metricKey]; ok {
		return collector.metrics[metric.Name][metricKey]
	}
	return 0
}
func (collector *MetricCollector) set(metric Metric, value float64) {
	var metricKey *Metric
	if val, ok := collector.metricKeys[(&metric).ToString()]; ok {
		metricKey = val
	} else {
		collector.metricKeys[(&metric).ToString()] = &metric
		metricKey = collector.metricKeys[(&metric).ToString()]
	}
	if _, ok := collector.metrics[metric.Name]; !ok {
		collector.metrics[metric.Name] = make(map[*Metric]float64, 0)
	}
	collector.metrics[metric.Name][metricKey] = value
}
