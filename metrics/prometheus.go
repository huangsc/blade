package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type prometheusMetrics struct {
	namespace string
	subsystem string
	labels    Labels
}

type prometheusCounter struct {
	counter prometheus.Counter
	vec     *prometheus.CounterVec
}

type prometheusGauge struct {
	gauge prometheus.Gauge
	vec   *prometheus.GaugeVec
}

type prometheusHistogram struct {
	histogram prometheus.Observer
	vec       *prometheus.HistogramVec
}

type prometheusSummary struct {
	summary prometheus.Observer
	vec     *prometheus.SummaryVec
}

// NewPrometheusMetrics 创建一个基于 Prometheus 的指标收集器
func NewPrometheusMetrics(opts ...Option) Metrics {
	options := &Options{
		Namespace: "",
		Subsystem: "",
		Buckets:   prometheus.DefBuckets,
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return &prometheusMetrics{
		namespace: options.Namespace,
		subsystem: options.Subsystem,
		labels:    options.ConstLabels,
	}
}

func (p *prometheusMetrics) Counter(name string, labels Labels) CounterMetric {
	opts := prometheus.CounterOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        name,
		ConstLabels: prometheus.Labels(p.labels),
	}

	if len(labels) == 0 {
		counter := prometheus.NewCounter(opts)
		prometheus.MustRegister(counter)
		return &prometheusCounter{counter: counter}
	}

	labelNames := make([]string, 0, len(labels))
	for name := range labels {
		labelNames = append(labelNames, name)
	}

	vec := prometheus.NewCounterVec(opts, labelNames)
	prometheus.MustRegister(vec)
	counter, err := vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusCounter{
		counter: counter,
		vec:     vec,
	}
}

func (p *prometheusMetrics) Gauge(name string, labels Labels) GaugeMetric {
	opts := prometheus.GaugeOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        name,
		ConstLabels: prometheus.Labels(p.labels),
	}

	if len(labels) == 0 {
		gauge := prometheus.NewGauge(opts)
		prometheus.MustRegister(gauge)
		return &prometheusGauge{gauge: gauge}
	}

	labelNames := make([]string, 0, len(labels))
	for name := range labels {
		labelNames = append(labelNames, name)
	}

	vec := prometheus.NewGaugeVec(opts, labelNames)
	prometheus.MustRegister(vec)
	gauge, err := vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusGauge{
		gauge: gauge,
		vec:   vec,
	}
}

func (p *prometheusMetrics) Histogram(name string, labels Labels) HistogramMetric {
	opts := prometheus.HistogramOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        name,
		ConstLabels: prometheus.Labels(p.labels),
	}

	if len(labels) == 0 {
		histogram := prometheus.NewHistogram(opts)
		prometheus.MustRegister(histogram)
		return &prometheusHistogram{histogram: histogram}
	}

	labelNames := make([]string, 0, len(labels))
	for name := range labels {
		labelNames = append(labelNames, name)
	}

	vec := prometheus.NewHistogramVec(opts, labelNames)
	prometheus.MustRegister(vec)
	histogram, err := vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusHistogram{
		histogram: histogram,
		vec:       vec,
	}
}

func (p *prometheusMetrics) Summary(name string, labels Labels) SummaryMetric {
	opts := prometheus.SummaryOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        name,
		ConstLabels: prometheus.Labels(p.labels),
	}

	if len(labels) == 0 {
		summary := prometheus.NewSummary(opts)
		prometheus.MustRegister(summary)
		return &prometheusSummary{summary: summary}
	}

	labelNames := make([]string, 0, len(labels))
	for name := range labels {
		labelNames = append(labelNames, name)
	}

	vec := prometheus.NewSummaryVec(opts, labelNames)
	prometheus.MustRegister(vec)
	summary, err := vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusSummary{
		summary: summary,
		vec:     vec,
	}
}

func (c *prometheusCounter) Inc() {
	c.counter.Inc()
}

func (c *prometheusCounter) Add(value float64) {
	c.counter.Add(value)
}

func (c *prometheusCounter) WithLabels(labels Labels) CounterMetric {
	if c.vec == nil {
		return c
	}

	counter, err := c.vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusCounter{counter: counter, vec: c.vec}
}

func (g *prometheusGauge) Set(value float64) {
	g.gauge.Set(value)
}

func (g *prometheusGauge) Inc() {
	g.gauge.Inc()
}

func (g *prometheusGauge) Dec() {
	g.gauge.Dec()
}

func (g *prometheusGauge) Add(value float64) {
	g.gauge.Add(value)
}

func (g *prometheusGauge) Sub(value float64) {
	g.gauge.Sub(value)
}

func (g *prometheusGauge) WithLabels(labels Labels) GaugeMetric {
	if g.vec == nil {
		return g
	}

	gauge, err := g.vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusGauge{gauge: gauge, vec: g.vec}
}

func (h *prometheusHistogram) Observe(value float64) {
	h.histogram.Observe(value)
}

func (h *prometheusHistogram) WithLabels(labels Labels) HistogramMetric {
	if h.vec == nil {
		return h
	}

	histogram, err := h.vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusHistogram{histogram: histogram, vec: h.vec}
}

func (s *prometheusSummary) Observe(value float64) {
	s.summary.Observe(value)
}

func (s *prometheusSummary) WithLabels(labels Labels) SummaryMetric {
	if s.vec == nil {
		return s
	}

	summary, err := s.vec.GetMetricWith(prometheus.Labels(labels))
	if err != nil {
		panic(err)
	}

	return &prometheusSummary{summary: summary, vec: s.vec}
}
