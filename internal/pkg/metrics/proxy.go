package metrics

import "github.com/prometheus/client_golang/prometheus"

type Proxy struct {
	success prometheus.Counter
	failed  prometheus.Counter
	errors  prometheus.Counter
}

// NewProxy creates new Proxy metrics collector.
func NewProxy() Proxy {
	return Proxy{
		success: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Subsystem: "requests",
			Name:      "success",
			Help:      "The count of successful proxied requests.",
		}),
		failed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Subsystem: "requests",
			Name:      "failed",
			Help:      "The count of unsuccessful proxied requests.",
		}),
		errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "proxy",
			Subsystem: "internal",
			Name:      "errors",
			Help:      "The count of internal proxying errors (including bad requests).",
		}),
	}
}

// IncrementSuccessful increments successful proxied requests counter.
func (w *Proxy) IncrementSuccessful() { w.success.Inc() }

// IncrementFailed increments unsuccessful proxied requests counter.
func (w *Proxy) IncrementFailed() { w.failed.Inc() }

// IncrementErrors increments internal proxying errors counter.
func (w *Proxy) IncrementErrors() { w.errors.Inc() }

// Register metrics with registerer.
func (w *Proxy) Register(reg prometheus.Registerer) error {
	if err := reg.Register(w.success); err != nil {
		return err
	}

	if err := reg.Register(w.failed); err != nil {
		return err
	}

	if err := reg.Register(w.errors); err != nil {
		return err
	}

	return nil
}
