package proxy_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/proxy"
)

type fakeMetric struct {
	success, failed, errors int
}

func (r *fakeMetric) IncrementSuccessful() { r.success++ }
func (r *fakeMetric) IncrementFailed()     { r.failed++ }
func (r *fakeMetric) IncrementErrors()     { r.errors++ }

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestHandler_ServeHTTPInputErrors(t *testing.T) {
	var noopHTTPClientMock httpClientFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     http.Header{},
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("NOOP"))),
		}, nil
	}

	cases := []struct {
		name           string
		giveRequest    func() *http.Request
		giveReqVars    map[string]string
		wantStatusCode int
		wantStrings    []string
	}{
		{
			name: "without url",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "", http.NoBody)

				return req
			},
			wantStatusCode: http.StatusInternalServerError,
			wantStrings:    []string{"cannot extract requested URI"},
		},
		{
			name: "empty url",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing", http.NoBody)

				return req
			},
			giveReqVars:    map[string]string{"uri": ""},
			wantStatusCode: http.StatusBadRequest,
			wantStrings:    []string{"empty request path"},
		},
		{
			name: "wrong url",
			giveRequest: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "http://testing", http.NoBody)

				return req
			},
			giveReqVars:    map[string]string{"uri": "http/f."},
			wantStatusCode: http.StatusBadRequest,
			wantStrings:    []string{"cannot build target URI"},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var (
				req     = tt.giveRequest()
				rr      = httptest.NewRecorder()
				m       = fakeMetric{}
				handler = proxy.NewHandler(context.Background(), noopHTTPClientMock, &m)
			)

			if tt.giveReqVars != nil {
				req = mux.SetURLVars(req, tt.giveReqVars)
			}

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)

			for _, s := range tt.wantStrings {
				assert.Contains(t, rr.Body.String(), s)
			}

			assert.Equal(t, 0, m.success)
			assert.Equal(t, 0, m.failed)
			assert.Equal(t, 1, m.errors)
		})
	}
}

func TestHandler_ServeHTTPSuccess(t *testing.T) {
	var (
		giveBody = []byte("aaa")
		req, _   = http.NewRequest(http.MethodPatch, "http://testing", bytes.NewReader(giveBody))
		rr       = httptest.NewRecorder()
		m        = fakeMetric{}
		executed bool
		client   httpClientFunc = func(req *http.Request) (*http.Response, error) {
			executed = true

			assert.Equal(t, "https://example.com/foo?foo=one&bar=two#hash", req.URL.String())
			assert.Equal(t, "foobar", req.Header.Get("foo"))
			assert.Equal(t, http.MethodPatch, req.Method)

			body, _ := ioutil.ReadAll(req.Body)

			assert.Equal(t, giveBody, body)

			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     http.Header{"response header": []string{"response value"}},
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("hell yeah!"))),
			}, nil
		}
		handler = proxy.NewHandler(context.Background(), client, &m)
	)

	req = mux.SetURLVars(req, map[string]string{"uri": "https/example.com/foo?foo=one&bar=two#hash"})
	req.Header.Set("foo", "foobar")

	handler.ServeHTTP(rr, req)

	assert.True(t, executed)

	assert.Equal(t, "response value", rr.Header().Get("response header"))
	assert.Equal(t, "hell yeah!", rr.Body.String())

	assert.Equal(t, 1, m.success)
	assert.Equal(t, 0, m.failed)
	assert.Equal(t, 0, m.errors)
}
