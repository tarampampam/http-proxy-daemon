package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers"

	"github.com/stretchr/testify/assert"
)

func TestNewHTMLErrorHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	handlers.NewHTMLErrorHandler(http.StatusNotFound).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")

	body := rr.Body.String()

	assert.Contains(t, body, "<html")
	assert.Contains(t, body, "Not Found")
	assert.Contains(t, body, "404")
}
