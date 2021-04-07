package metrics_test

import (
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestProxy_Register(t *testing.T) {
	var (
		registry = prometheus.NewRegistry()
		p        = metrics.NewProxy()
	)

	assert.NoError(t, p.Register(registry))

	count, err := testutil.GatherAndCount(registry,
		"proxy_requests_success",
		"proxy_requests_failed",
		"proxy_internal_errors",
	)
	assert.NoError(t, err)

	assert.Equal(t, 3, count)
}

func TestProxy_IncrementSuccessful(t *testing.T) {
	p := metrics.NewProxy()

	p.IncrementSuccessful()

	metric := getMetric(t, &p, "proxy_requests_success")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestProxy_IncrementFailed(t *testing.T) {
	p := metrics.NewProxy()

	p.IncrementFailed()

	metric := getMetric(t, &p, "proxy_requests_failed")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestProxy_IncrementErrors(t *testing.T) {
	p := metrics.NewProxy()

	p.IncrementErrors()

	metric := getMetric(t, &p, "proxy_internal_errors")
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

type registerer interface {
	Register(prometheus.Registerer) error
}

func getMetric(t *testing.T, reg registerer, name string) *dto.Metric {
	t.Helper()

	registry := prometheus.NewRegistry()
	_ = reg.Register(registry)

	families, _ := registry.Gather()

	for _, family := range families {
		if family.GetName() == name {
			return family.Metric[0]
		}
	}

	assert.FailNowf(t, "cannot resolve metric for: %s", name)

	return nil
}
