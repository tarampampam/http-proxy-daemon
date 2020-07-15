package index

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandler_ServeHTTPWithoutRoutes(t *testing.T) {
	t.Parallel()

	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	NewHandler(mux.NewRouter()).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `[]`, rr.Body.String())
}

func TestHandler_ServeHTTPWithRoutes(t *testing.T) {
	t.Parallel()

	var (
		req, _  = http.NewRequest("GET", "http://testing", nil)
		rr      = httptest.NewRecorder()
		router  = mux.NewRouter()
		handler = NewHandler(router)
	)

	router.
		HandleFunc("/foo", func(http.ResponseWriter, *http.Request) {}).
		Methods(http.MethodGet).
		Name("foo1")

	router.
		HandleFunc("/bar", func(http.ResponseWriter, *http.Request) {}).
		Methods(http.MethodPost).
		Name("bar2")

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `["/foo","/bar"]`, rr.Body.String())
}

func TestHandler_ServeHTTPWithRouterError(t *testing.T) {
	t.Parallel()

	var (
		req, _  = http.NewRequest("GET", "http://testing", nil)
		rr      = httptest.NewRecorder()
		router  = mux.NewRouter()
		handler = NewHandler(router)
	)

	router.
		//HandleFunc("/foo", func(http.ResponseWriter, *http.Request) {}).
		Methods(http.MethodGet).
		Name("foo1")

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":500, "error":true, "message":"mux: route doesn't have a path"}`, rr.Body.String())
}
