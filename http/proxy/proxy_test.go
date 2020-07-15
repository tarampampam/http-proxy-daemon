package proxy

import (
	"bytes"
	"errors"
	"http-proxy-daemon/counters"
	"http-proxy-daemon/shared"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn), //nolint:unconvert
	}
}

type httpTimeoutError struct {
	e       error
	timeout bool
}

func (e *httpTimeoutError) Timeout() bool { return e.timeout }
func (e *httpTimeoutError) Error() string { return e.e.Error() }

func TestHandler_ServeHTTPWrongPathRequested(t *testing.T) {
	t.Parallel()

	var (
		req, _ = http.NewRequest("GET", "http://foobar", nil)
		rr     = httptest.NewRecorder()
		c      = counters.NewInMemoryCounters(nil)
	)

	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))

	NewHandler(c, time.Second*1, 1).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":500, "error":true, "message":"Cannot extract requested path"}`, rr.Body.String())

	assert.Equal(t, int64(1), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))
}
func TestHandler_ServeHTTPEmptyPathRequested(t *testing.T) {
	t.Parallel()

	var (
		req, _      = http.NewRequest("GET", "http://foobar", nil)
		reqWithVars = mux.SetURLVars(req, map[string]string{"uri": "foobar"})
		rr          = httptest.NewRecorder()
		c           = counters.NewInMemoryCounters(nil)
	)

	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))

	NewHandler(c, time.Second*1, 1).ServeHTTP(rr, reqWithVars)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":400, "error":true, "message":"Empty request path"}`, rr.Body.String())

	assert.Equal(t, int64(1), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))
}

func TestHandler_ServeHTTPTimeoutError(t *testing.T) {
	t.Parallel()

	var (
		req, _      = http.NewRequest("GET", "http://testing", nil)
		reqWithVars = mux.SetURLVars(req, map[string]string{"uri": "http/foo.com/bar?baz=blah"})
		rr          = httptest.NewRecorder()
		c           = counters.NewInMemoryCounters(nil)
		handler     = NewHandler(c, time.Second*1, 1)
	)

	handler.(*Handler).httpClient = newTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "http://foo.com/bar?baz=blah", req.URL.String())

		return nil, &httpTimeoutError{
			e:       errors.New("timeout error"),
			timeout: true,
		}
	})

	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))

	handler.ServeHTTP(rr, reqWithVars)

	assert.Equal(t, http.StatusRequestTimeout, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":408, "error":true, "message":"Request timeout exceeded"}`, rr.Body.String())

	assert.Equal(t, int64(1), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))
}

func TestHandler_ServeHTTPRegularUsage(t *testing.T) {
	t.Parallel()

	var (
		req, _      = http.NewRequest("GET", "http://testing", nil)
		reqWithVars = mux.SetURLVars(req, map[string]string{"uri": "http/foo.com/bar?baz=blah"})
		rr          = httptest.NewRecorder()
		c           = counters.NewInMemoryCounters(nil)
		handler     = NewHandler(c, time.Second*1, 1)
	)

	reqWithVars.Header.Add("Aaa", "bbb")

	handler.(*Handler).httpClient = newTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "http://foo.com/bar?baz=blah", req.URL.String())
		assert.Equal(t, "bbb", req.Header.Get("Aaa"))

		headers := http.Header{}
		headers.Add("foo", "bar")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("Fake response")),
			Header:     headers,
		}, nil
	})

	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedSuccess))

	handler.ServeHTTP(rr, reqWithVars)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "bar", rr.Header().Get("foo"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Fake response", rr.Body.String())

	assert.Equal(t, int64(0), c.Get(shared.MetricProxiedErrors))
	assert.Equal(t, int64(1), c.Get(shared.MetricProxiedSuccess))
}
