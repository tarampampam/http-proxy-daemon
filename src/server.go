package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type IServer interface {
	RegisterHandlers()
	Start() error
}

// Proxy server structure.
type Server struct {
	server           *http.Server
	router           *mux.Router
	client           *http.Client
	proxyRoutePrefix string
	stdLog           *log.Logger
	errLog           *log.Logger
	counters         ICounter
	startTime        time.Time
}

const (
	metricProxiedSuccess = "proxied_success"
	metricProxiedErrors  = "proxied_errors"
)

// Server constructor.
func NewServer(host string, port int, proxyPrefix string, stdLog, errLog *log.Logger) *Server {
	var router = *mux.NewRouter()
	var tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // skip ssl errors
	}

	return &Server{
		server: &http.Server{
			Addr:     host + ":" + strconv.Itoa(port), // TCP address and port to listen on
			Handler:  &router,
			ErrorLog: errLog,
		},
		router:           &router,
		proxyRoutePrefix: proxyPrefix,
		client: &http.Client{
			Transport: tr,
			Timeout:   time.Second * 30, // Set request timeout
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return errors.New("request: too many (5) redirects")
				}
				return nil
			},
		},
		stdLog:   stdLog,
		counters: NewCounters(nil),
	}
}

// Register server http handlers.
func (s *Server) RegisterHandlers() {
	s.router.HandleFunc("/", s.indexHandler).
		Methods("GET").
		Name("index")
	s.router.HandleFunc("/ping", s.pingHandler).
		Methods("GET").
		Name("ping")
	s.router.HandleFunc("/metrics", s.metricsHandler).
		Methods("GET").
		Name("metrics")
	s.router.HandleFunc("/"+s.proxyRoutePrefix+"/{uri:.*}", s.proxyRequestHandler).
		Methods("GET", "POST", "HEAD", "PUT", "PATCH", "DELETE", "OPTIONS").
		Name("proxy")

	s.router.NotFoundHandler = s.notFoundHandler()
	s.router.MethodNotAllowedHandler = s.methodNotAllowedHandler()
}

// Start proxy server.
func (s *Server) Start() error {
	s.startTime = time.Now()
	s.stdLog.Println("Starting server on", s.server.Addr)
	return s.server.ListenAndServe()
}

// Start TSL proxy server.
func (s *Server) StartSSL(certFile, keyFile string) error {
	s.startTime = time.Now()
	s.stdLog.Println("Starting TSL server on", s.server.Addr)
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

// Stop proxy server.
func (s *Server) Stop() error {
	s.stdLog.Println("Stopping server")
	return s.server.Shutdown(context.Background())
}

// Set http client response timeout.
func (s *Server) SetClientResponseTimeout(time time.Duration) {
	s.client.Timeout = time
}

// Error handler - 404
func (s *Server) notFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.errorHandler(w, *NewServerError(http.StatusNotFound, "Not found"))
	})
}

// Error handler - 405
func (s *Server) methodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.errorHandler(w, *NewServerError(http.StatusMethodNotAllowed, "Method not allowed"))
	})
}

// Our custom http server errors handler (should be called manually).
func (s *Server) errorHandler(w http.ResponseWriter, error ServerError) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(error.Code)
	_ = json.NewEncoder(w).Encode(error)
}

// Index route handler. Show all available routes in a json response.
func (s *Server) indexHandler(w http.ResponseWriter, _ *http.Request) {
	var routes []string
	// Walk through all available routes an fill routes slice
	err := s.router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		if t, err := route.GetPathTemplate(); err == nil {
			routes = append(routes, t)
			return nil
		} else {
			return err
		}
	})
	// Handle possible error
	if err != nil {
		s.errorHandler(w, *NewServerError(http.StatusBadRequest, err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	// And print results
	_ = json.NewEncoder(w).Encode(routes)
}

// Ping request handler
func (s *Server) pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("pong")
}

// Metrics request handler.
func (s *Server) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	res := make(map[string]interface{})
	// Append metric proxy stats
	res["proxied_success"] = s.counters.Get(metricProxiedSuccess)
	res["proxied_errors"] = s.counters.Get(metricProxiedErrors)
	// Append uptime in seconds
	res["uptime_sec"] = int64(time.Since(s.startTime).Seconds())
	// Append hostname
	if h, err := os.Hostname(); err == nil {
		res["hostname"] = h
	}
	// Append version
	res["version"] = VERSION

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(res)
}

// Proxy request handler.
func (s *Server) proxyRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Increment counter value at starts (decrement it later if all is ok)
	s.counters.Increment(metricProxiedErrors)

	var deferredLogger = NewDeferredLogger(s.stdLog)
	deferredLogger.Add(fmt.Sprintf(`[%s "%s"] - "%s %s"`, r.RemoteAddr, r.UserAgent(), r.Method, r.URL.String()))
	defer deferredLogger.Flush(" - ")

	// Make sure that "uri" are presents
	uri, uriFound := mux.Vars(r)["uri"]
	if !uriFound {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot extract requested path"))
		return
	}

	// Extract request schema and path from route
	var schema, path = s.uriToSchemaAndPath(uri)
	if len(path) == 0 {
		s.errorHandler(w, *NewServerError(http.StatusBadRequest, "Empty request path"))
		return
	}

	// Build target uri
	var target = s.buildTargetUri(schema, path, r.URL.RawQuery)
	deferredLogger.Add(fmt.Sprintf("<%s>", target))

	// Create HTTP request
	httpRequest, reqErr := http.NewRequest(r.Method, target, nil)
	if reqErr != nil {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot prepare http request: "+reqErr.Error()))
		return
	}

	// Proxy all request headers
	httpRequest.Header = r.Header.Clone()

	// Make an http request
	resp, respErr := s.client.Do(httpRequest)

	// Check for response error
	if respErr != nil {
		deferredLogger.Add(fmt.Sprintf(`ERROR "%s"`, respErr.Error()))
		if e, ok := respErr.(*url.Error); ok {
			if e.Timeout() {
				s.errorHandler(w, *NewServerError(http.StatusRequestTimeout, "Request timeout exceeded"))
				return
			}
		}
		s.errorHandler(w, *NewServerError(http.StatusServiceUnavailable, respErr.Error()))
		return
	}

	// Check for response
	if resp == nil {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Response not received"))
		return
	}

	// Close response body after all
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}(resp)

	// Write headers
	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ";"))
	}

	// Allow access from anywhere
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Write response code
	w.WriteHeader(resp.StatusCode)

	// Write content
	responseLen, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot write response: "+copyErr.Error()))
		return
	}

	deferredLogger.Add(fmt.Sprintf(`"HTTP %d (%d bytes)"`, resp.StatusCode, responseLen))

	s.counters.Decrement(metricProxiedErrors) // If all is ok - decrement counter value
	s.counters.Increment(metricProxiedSuccess)
}

// Extract schema and path from passed specific uri string.
func (s *Server) uriToSchemaAndPath(uri string) (string, string) {
	var schema, path = "", uri
	// Try to extract "schema" part (substring)
	if strings.Contains(uri, "/") {
		// Extract value
		slashPos := strings.IndexByte(uri, '/')
		possibleSchema := strings.ToLower(uri[:slashPos])
		// Validate extracted schema and set
		if s.validateHttpSchema(possibleSchema) {
			schema, path = possibleSchema, uri[slashPos+1:]
		}
	}
	return schema, path
}

// Target uri builder.
func (*Server) buildTargetUri(schema, domainAndPath, params string) (uri string) {
	// Write schema
	if len(schema) != 0 {
		uri += schema
	} else {
		uri += "http"
	}
	// Write domain and path
	uri += "://" + domainAndPath
	// Write query params
	if len(params) != 0 {
		uri += "?" + params
	}
	// Cannot be less then..
	if len(uri) < 8 {
		return ""
	}
	return uri
}

// Schema validator.
func (*Server) validateHttpSchema(schema string) bool {
	if l := len(schema); l < 4 || l > 6 {
		return false // fast check
	}
	// Valid (allowed) schemas list
	var allowedSchemas = [...]string{"http", "https"}
	// Try to find passed schema in allowed schemas list
	for _, a := range allowedSchemas {
		if a == schema {
			return true
		}
	}
	return false
}
