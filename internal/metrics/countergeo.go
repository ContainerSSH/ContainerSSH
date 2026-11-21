package metrics

import (
	"net"
)

type counterGeoImpl struct {
	name      string
	collector *collector
	labels    []MetricLabel
}

func (c *counterGeoImpl) Increment(ip net.IP, labels ...MetricLabel) {
	_ = c.IncrementBy(ip, 1, labels...)
}

func (c *counterGeoImpl) IncrementBy(ip net.IP, by float64, labels ...MetricLabel) error {
	c.collector.mutex.Lock()
	defer c.collector.mutex.Unlock()

	if by < 0 {
		return CounterCannotBeIncrementedByNegative
	}

	realLabels := metricLabels(
		append(c.labels, labels...),
	).toMap()
	realLabels["country"] = c.collector.geoIpLookupProvider.Lookup(ip)

	value := c.collector.get(c.name, realLabels)
	c.collector.set(c.name, realLabels, value+by)
	return nil
}

func (c *counterGeoImpl) WithLabels(labels ...MetricLabel) GeoCounter {
	return &counterGeoImpl{
		name:      c.name,
		collector: c.collector,
		labels:    append(c.labels, labels...),
	}
}
