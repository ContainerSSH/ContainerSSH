package metrics

import (
	"github.com/janoszen/containerssh/geoip"
	"net"
	"sort"
	"sync"
)

type MetricCollector struct {
	mutex               *sync.Mutex
	metricKeys          map[string]*Metric
	metrics             map[string]map[string]float64
	help                map[string]string
	types               map[string]MetricType
	geoIpLookupProvider geoip.LookupProvider
}

func New(geoIpLookupProvider geoip.LookupProvider) *MetricCollector {
	return &MetricCollector{
		&sync.Mutex{},
		make(map[string]*Metric, 0),
		make(map[string]map[string]float64, 0),
		make(map[string]string, 0),
		make(map[string]MetricType, 0),
		geoIpLookupProvider,
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
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}

func (collector *MetricCollector) GetMetrics(name string) map[string]float64 {
	if val, ok := collector.metrics[name]; ok {
		return val
	} else {
		return make(map[string]float64, 0)
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

func (collector *MetricCollector) IncrementGeo(metric Metric, remoteAddr net.IP) float64 {
	metric.Labels["country"] = collector.geoIpLookupProvider.Lookup(remoteAddr)
	return collector.Increment(metric)
}

func (collector *MetricCollector) Decrement(metric Metric) float64 {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	value := collector.get(metric)
	value = value - 1
	collector.set(metric, value)
	return value
}

func (collector *MetricCollector) DecrementGeo(metric Metric, remoteAddr net.IP) float64 {
	metric.Labels["country"] = collector.geoIpLookupProvider.Lookup(remoteAddr)
	return collector.Decrement(metric)
}

func (collector *MetricCollector) get(metric Metric) float64 {
	if _, ok := collector.metricKeys[(&metric).ToString()]; !ok {
		collector.metricKeys[(&metric).ToString()] = &metric
	}
	if _, ok := collector.metrics[metric.Name]; !ok {
		collector.metrics[metric.Name] = make(map[string]float64, 0)
		return 0
	}
	if _, ok := collector.metrics[metric.Name][(&metric).ToString()]; ok {
		return collector.metrics[metric.Name][(&metric).ToString()]
	}
	return 0
}
func (collector *MetricCollector) set(metric Metric, value float64) {
	if _, ok := collector.metricKeys[(&metric).ToString()]; !ok {
		collector.metricKeys[(&metric).ToString()] = &metric
	}
	if _, ok := collector.metrics[metric.Name]; !ok {
		collector.metrics[metric.Name] = make(map[string]float64, 0)
	}
	collector.metrics[metric.Name][(&metric).ToString()] = value
}
