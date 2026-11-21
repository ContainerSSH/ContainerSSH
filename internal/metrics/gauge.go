package metrics

type gaugeImpl struct {
	name      string
	collector *collector
	labels    []MetricLabel
}

func (g *gaugeImpl) Increment(labels ...MetricLabel) {
	g.IncrementBy(1, labels...)
}

func (g *gaugeImpl) IncrementBy(by float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	value := g.collector.get(g.name, realLabels)
	g.collector.set(g.name, realLabels, value+by)
}

func (g *gaugeImpl) Decrement(labels ...MetricLabel) {
	g.DecrementBy(1, labels...)
}

func (g *gaugeImpl) DecrementBy(by float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	value := g.collector.get(g.name, realLabels)
	g.collector.set(g.name, realLabels, value-by)
}

func (g *gaugeImpl) Set(value float64, labels ...MetricLabel) {
	g.collector.mutex.Lock()
	defer g.collector.mutex.Unlock()

	realLabels := metricLabels(
		append(g.labels, labels...),
	).toMap()
	g.collector.set(g.name, realLabels, value)
}

func (g *gaugeImpl) WithLabels(labels ...MetricLabel) Gauge {
	return &gaugeImpl{
		name:      g.name,
		collector: g.collector,
		labels:    append(g.labels, labels...),
	}
}
