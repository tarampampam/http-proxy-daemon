package index_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/index"
)

func TestNewHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	index.NewHandler().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")

	body := rr.Body.String()

	assert.Contains(t, body, "<html")
	assert.Contains(t, body, "HTTP Proxy Daemon")
}
