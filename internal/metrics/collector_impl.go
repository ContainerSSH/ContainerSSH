package metrics

import (
	"bytes"
	"sync"

	"github.com/containerssh/libcontainerssh/internal/geoip/geoipprovider"
)

type collector struct {
	geoIpLookupProvider geoipprovider.LookupProvider

	mutex      *sync.Mutex
	metricsMap map[string]Metric
	metrics    []Metric
	values     map[string]*metricValue
}

func (c *collector) MustCreateCounter(name string, unit string, help string) Counter {
	counter, err := c.CreateCounter(name, unit, help)
	if err != nil {
		panic(err)
	}
	return counter
}

func (c *collector) MustCreateCounterGeo(name string, unit string, help string) GeoCounter {
	counter, err := c.CreateCounterGeo(name, unit, help)
	if err != nil {
		panic(err)
	}
	return counter
}

func (c *collector) MustCreateGauge(name string, unit string, help string) Gauge {
	gauge, err := c.CreateGauge(name, unit, help)
	if err != nil {
		panic(err)
	}
	return gauge
}

func (c *collector) MustCreateGaugeGeo(name string, unit string, help string) GeoGauge {
	gauge, err := c.CreateGaugeGeo(name, unit, help)
	if err != nil {
		panic(err)
	}
	return gauge
}

type metricValue struct {
	metricValueMap map[string]*MetricValue
	metricValues   []*MetricValue
}

func (c *collector) createMetric(name string, unit string, help string, metricType MetricType) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if existingMetric, ok := c.metricsMap[name]; ok {
		if existingMetric.Name == name &&
			existingMetric.Unit == unit &&
			existingMetric.Help == help &&
			existingMetric.Type == metricType {
			return nil
		}
		return MetricAlreadyExists
	}
	metric := Metric{
		Name: name,
		Unit: unit,
		Help: help,
		Type: metricType,
	}

	c.metricsMap[name] = metric
	c.metrics = append(c.metrics, metric)
	return nil
}

func (c *collector) CreateCounter(name string, unit string, help string) (Counter, error) {
	if err := c.createMetric(name, unit, help, MetricTypeCounter); err != nil {
		return nil, err
	}
	return &counterImpl{
		name:      name,
		collector: c,
	}, nil
}

func (c *collector) CreateCounterGeo(name string, unit string, help string) (GeoCounter, error) {
	if err := c.createMetric(name, unit, help, MetricTypeCounter); err != nil {
		return nil, err
	}
	return &counterGeoImpl{
		name:      name,
		collector: c,
	}, nil
}

func (c *collector) CreateGauge(name string, unit string, help string) (Gauge, error) {
	if err := c.createMetric(name, unit, help, MetricTypeGauge); err != nil {
		return nil, err
	}
	return &gaugeImpl{
		name:      name,
		collector: c,
	}, nil
}

func (c *collector) CreateGaugeGeo(name string, unit string, help string) (GeoGauge, error) {
	if err := c.createMetric(name, unit, help, MetricTypeGauge); err != nil {
		return nil, err
	}
	return &gaugeGeoImpl{
		name:      name,
		collector: c,
	}, nil
}

func (c *collector) ListMetrics() []Metric {
	return c.metrics
}

func (c *collector) GetMetric(name string) []MetricValue {
	var results []MetricValue
	if val, ok := c.values[name]; ok {
		for _, v := range val.metricValues {
			results = append(results, *v)
		}
	}
	return results
}

func (c *collector) String() string {
	var buffer bytes.Buffer
	for _, metric := range c.metrics {
		buffer.WriteString(metric.String())
		for _, value := range c.GetMetric(metric.Name) {
			buffer.WriteString(value.String())
		}
	}
	buffer.WriteString("# EOF\n")
	return buffer.String()
}

func (c *collector) getValueStruct(name string, labels map[string]string) *MetricValue {
	if _, ok := c.values[name]; !ok {
		c.values[name] = &metricValue{
			metricValueMap: map[string]*MetricValue{},
			metricValues:   []*MetricValue{},
		}
	}
	valueStruct := MetricValue{
		Name:   name,
		Labels: labels,
		Value:  0,
	}
	if _, ok2 := c.values[name].metricValueMap[valueStruct.CombinedName()]; !ok2 {
		c.values[name].metricValueMap[valueStruct.CombinedName()] = &valueStruct
		c.values[name].metricValues = append(c.values[name].metricValues, &valueStruct)
	}
	return c.values[name].metricValueMap[valueStruct.CombinedName()]
}

func (c *collector) get(name string, labels map[string]string) float64 {
	return c.getValueStruct(name, labels).Value
}

func (c *collector) set(name string, labels map[string]string, value float64) {
	c.getValueStruct(name, labels).Value = value
}
