package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundHandler(t *testing.T) {
	t.Parallel()

	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	NotFoundHandler().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":404,"error":true,"message":"Not found"}`, rr.Body.String())
}

func TestMethodNotAllowedHandler(t *testing.T) {
	t.Parallel()

	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	MethodNotAllowedHandler().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":405,"error":true,"message":"Method not allowed"}`, rr.Body.String())
}
