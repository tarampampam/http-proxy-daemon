package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/tarampampam/http-proxy-daemon/internal/pkg/config"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/handlers"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/middlewares/logreq"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/http/middlewares/panic"
	"github.com/tarampampam/http-proxy-daemon/internal/pkg/metrics"
)

type (
	Server struct {
		log    *zap.Logger
		server *http.Server
		router *mux.Router
	}
)

const (
	readTimeout  = time.Second * 3
	writeTimeout = time.Second * 60 // this is maximal proxy response timeout also
)

// NewServer creates new server instance.
func NewServer(log *zap.Logger) *Server {
	var (
		router     = mux.NewRouter()
		httpServer = &http.Server{
			Handler:           router,
			ErrorLog:          zap.NewStdLog(log),
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			ReadHeaderTimeout: readTimeout,
		}
	)

	return &Server{
		log:    log,
		server: httpServer,
		router: router,
	}
}

// Register server routes, middlewares, etc.
func (s *Server) Register(ctx context.Context, cfg config.Config) error {
	registry := metrics.NewRegistry()

	s.registerGlobalMiddlewares()

	return s.registerHandlers(ctx, cfg, registry)
}

func (s *Server) registerGlobalMiddlewares() {
	s.router.Use(
		logreq.New(s.log),
		panic.New(s.log),
	)
}

// registerHandlers register server http handlers.
func (s *Server) registerHandlers(ctx context.Context, cfg config.Config, registry *prometheus.Registry) error {
	s.router.NotFoundHandler = handlers.NewHTMLErrorHandler(http.StatusNotFound)
	s.router.MethodNotAllowedHandler = handlers.NewHTMLErrorHandler(http.StatusMethodNotAllowed)

	if err := s.registerProxyRoutes(ctx, cfg, registry); err != nil {
		return err
	}

	s.registerIndexHandler()
	s.registerServiceHandlers(registry)

	return nil
}

// Start server.
func (s *Server) Start(ip string, port uint16) error {
	s.server.Addr = ip + ":" + strconv.Itoa(int(port))

	return s.server.ListenAndServe()
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error { return s.server.Shutdown(ctx) }
