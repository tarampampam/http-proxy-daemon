package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_constants(t *testing.T) {
	assert.Equal(t, "proxied_success", MetricProxiedSuccess)
	assert.Equal(t, "proxied_errors", MetricProxiedErrors)
}
