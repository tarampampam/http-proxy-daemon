package metrics

import (
	"encoding/json"
	"http-proxy-daemon/shared"
	"http-proxy-daemon/version"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeCounters struct{}

func (f fakeCounters) Increment(name string)        { panic("wrong usage") }
func (f fakeCounters) Decrement(name string)        { panic("wrong usage") }
func (f fakeCounters) Set(name string, value int64) { panic("wrong usage") }
func (f fakeCounters) Get(name string) int64 {
	switch name {
	case shared.MetricProxiedErrors:
		return 555
	case shared.MetricProxiedSuccess:
		return 666
	}

	panic("wrong usage")
}

func TestHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	var (
		req, _             = http.NewRequest("GET", "http://testing", nil)
		rr                 = httptest.NewRecorder()
		serverStartTime    = time.Now().Add(-time.Second * 10)
		currentHostname, _ = os.Hostname()
	)

	NewHandler(&serverStartTime, fakeCounters{}).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	data := make(map[string]interface{})
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, currentHostname, data["hostname"].(string))
	assert.Equal(t, float64(555), data["proxied_errors"].(float64))
	assert.Equal(t, float64(666), data["proxied_success"].(float64))
	assert.Equal(t, float64(10), data["uptime_sec"].(float64))
	assert.Equal(t, version.Version(), data["version"].(string))
}
