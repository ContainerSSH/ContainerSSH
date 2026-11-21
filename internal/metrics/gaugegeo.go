package metrics

import (
	"net"
)

type gaugeGeoImpl struct {
	name      string
	collector *collector
	labels    []MetricLabel
}

func (g *gaugeGeoImpl) Increment(ip net.IP, labels ...MetricLabel) {
	g.IncrementBy(ip, 1, labels...)
}

func (g *gaugeGeoImpl) IncrementBy(ip net.IP, by float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	realLabels["country"] = g.collector.geoIpLookupProvider.Lookup(ip)

	value := g.collector.get(g.name, realLabels)
	g.collector.set(g.name, realLabels, value+by)
}

func (g *gaugeGeoImpl) Decrement(ip net.IP, labels ...MetricLabel) {
	g.DecrementBy(ip, 1, labels...)
}

func (g *gaugeGeoImpl) DecrementBy(ip net.IP, by float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	realLabels["country"] = g.collector.geoIpLookupProvider.Lookup(ip)

	value := g.collector.get(g.name, realLabels)
	g.collector.set(g.name, realLabels, value-by)
}

func (g *gaugeGeoImpl) Set(ip net.IP, value float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	realLabels["country"] = g.collector.geoIpLookupProvider.Lookup(ip)

	g.collector.set(g.name, realLabels, value)
}

func (g *gaugeGeoImpl) WithLabels(labels ...MetricLabel) GeoGauge {
	return &gaugeGeoImpl{
		name:      g.name,
		collector: g.collector,
		labels:    append(g.labels, labels...),
	}
}
