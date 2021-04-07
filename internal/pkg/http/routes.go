package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/checkers"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/config"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/healthz"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/index"
	metricsHandler "github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/metrics"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers/proxy"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/metrics"
)

func (s *Server) registerProxyRoutes(ctx context.Context, cfg config.Config, registerer prometheus.Registerer) error {
	if cfg.Proxy.Prefix == "" {
		return errors.New("empty proxy prefix")
	}

	proxyMetrics := metrics.NewProxy()
	if err := proxyMetrics.Register(registerer); err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec //lgtm [go/disabled-certificate-check]
			},
		},
		Timeout: cfg.Proxy.RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			const maxRedirects int = 3

			if len(via) >= maxRedirects {
				return errors.New("too many (" + strconv.Itoa(maxRedirects) + ") redirects")
			}

			return nil
		},
	}

	s.router.
		Handle("/"+cfg.Proxy.Prefix+"/{uri:.*}", proxy.NewHandler(ctx, httpClient, &proxyMetrics)).
		Name("proxy")

	return nil
}

func (s *Server) registerIndexHandler() {
	s.router.
		Handle("/", index.NewHandler()).
		Methods(http.MethodGet).
		Name("index")
}

func (s *Server) registerServiceHandlers(registry prometheus.Gatherer) {
	s.router.
		HandleFunc("/metrics", metricsHandler.NewHandler(registry)).
		Methods(http.MethodGet).
		Name("metrics")

	s.router.
		HandleFunc("/ready", healthz.NewHandler(checkers.NewReadyChecker())).
		Methods(http.MethodGet, http.MethodHead).
		Name("ready")

	s.router.
		HandleFunc("/live", healthz.NewHandler(checkers.NewLiveChecker())).
		Methods(http.MethodGet, http.MethodHead).
		Name("live")
}
