package errors

import "net/http"

// Error handler - 404
func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(NewServerError(http.StatusNotFound, "Not found").ToJSON())
	})
}

// Error handler - 405
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write(NewServerError(http.StatusMethodNotAllowed, "Method not allowed").ToJSON())
	})
}
