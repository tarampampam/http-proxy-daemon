package http

import (
	"context"
	"http-proxy-daemon/counters"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type (
	ServerSettings struct {
		Address             string // TCP address to listen on
		ProxyRoutePrefix    string
		WriteTimeout        time.Duration
		ReadTimeout         time.Duration
		ProxyRequestTimeout time.Duration
		KeepAliveEnabled    bool
	}

	Server struct {
		settings  *ServerSettings
		Server    *http.Server
		Router    *mux.Router
		startTime time.Time
		counters  counters.Counters
		stdLog    *log.Logger
		errLog    *log.Logger
	}
)

// NewServer creates new server instance.
func NewServer(settings *ServerSettings) *Server {
	var (
		router     = *mux.NewRouter()
		stdLog     = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
		errLog     = log.New(os.Stderr, "[error] ", log.LstdFlags)
		httpServer = &http.Server{
			Addr:         settings.Address,
			Handler:      handlers.CombinedLoggingHandler(os.Stdout, &router),
			ErrorLog:     errLog,
			WriteTimeout: settings.WriteTimeout,
			ReadTimeout:  settings.ReadTimeout,
		}
	)

	httpServer.SetKeepAlivesEnabled(settings.KeepAliveEnabled)

	return &Server{
		settings: settings,
		Server:   httpServer,
		Router:   &router,
		counters: counters.NewInMemoryCounters(nil),
		stdLog:   stdLog,
		errLog:   errLog,
	}
}

// Start proxy Server.
func (s *Server) Start() error {
	s.startTime = time.Now()
	s.stdLog.Println("Starting Server on " + s.Server.Addr)

	return s.Server.ListenAndServe()
}

// Stop proxy Server.
func (s *Server) Stop(ctx context.Context) error {
	s.stdLog.Println("Stopping Server")

	return s.Server.Shutdown(ctx)
}
