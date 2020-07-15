package http

import (
	"http-proxy-daemon/http/errors"
	"http-proxy-daemon/http/index"
	"http-proxy-daemon/http/metrics"
	"http-proxy-daemon/http/ping"
	"http-proxy-daemon/http/proxy"
	"net/http"
	"strings"
	"time"
)

// HTTP request timeout
const (
	httpRequestTimeout time.Duration = time.Second * 30
	maxRedirects       int           = 2
)

// RegisterHandlers register server http handlers.
func (s *Server) RegisterHandlers() {
	s.Router.NotFoundHandler = errors.NotFoundHandler()
	s.Router.MethodNotAllowedHandler = errors.MethodNotAllowedHandler()

	s.Router.
		Handle("/", index.NewHandler(s.Router)).
		Methods(http.MethodGet).
		Name("index")

	s.Router.
		Handle("/ping", DisableCachingMiddleware(ping.NewHandler())).
		Methods(http.MethodGet).
		Name("ping")

	s.Router.
		Handle("/metrics", DisableCachingMiddleware(metrics.NewHandler(&s.startTime, s.counters))).
		Methods(http.MethodGet).
		Name("metrics")

	s.Router.
		Handle(
			"/"+strings.TrimLeft(s.Settings.ProxyRoutePrefix+"/{uri:.*}", "/"),
			proxy.NewHandler(s.counters, httpRequestTimeout, maxRedirects),
		).
		Methods(
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		).
		Name("proxy")
}
