package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	defaultRegistry = prometheus.NewRegistry()
)

// Counter is a Prometheus counter metric.
type Counter struct {
	counter prometheus.Counter
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.counter.Inc()
}

// Add adds the given value to the counter.
func (c *Counter) Add(val float64) {
	c.counter.Add(val)
}

// NewCounter creates a new Counter metric.
func NewCounter(name, help string) Counter {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
	defaultRegistry.MustRegister(c)
	return Counter{counter: c}
}

// Gauge is a Prometheus gauge metric.
type Gauge struct {
	gauge prometheus.Gauge
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(val float64) {
	g.gauge.Set(val)
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.gauge.Inc()
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.gauge.Dec()
}

// NewGauge creates a new Gauge metric.
func NewGauge(name, help string) Gauge {
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
	defaultRegistry.MustRegister(g)
	return Gauge{gauge: g}
}

// Histogram is a Prometheus histogram metric.
type Histogram struct {
	histogram prometheus.Histogram
}

// Observe records an observation in the histogram.
func (h *Histogram) Observe(val float64) {
	h.histogram.Observe(val)
}

// NewHistogram creates a new Histogram metric.
func NewHistogram(name, help string, buckets []float64) Histogram {
	h := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	})
	defaultRegistry.MustRegister(h)
	return Histogram{histogram: h}
}

// Handler returns a Gin handler that exposes metrics in Prometheus format.
func Handler() gin.HandlerFunc {
	h := promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
