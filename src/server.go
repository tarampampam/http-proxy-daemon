package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Proxy server structure.
type Server struct {
	server *http.Server
	router *mux.Router
	client *http.Client
	log    *log.Logger
}

// Server constructor.
func NewServer(host, port string, log *log.Logger) *Server {
	var router = *mux.NewRouter()
	var tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // skip ssl errors
	}

	return &Server{
		server: &http.Server{
			Addr:    host + ":" + port, // TCP address to listen on, ":http" if empty
			Handler: &router,
		},
		router: &router,
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
		log: log,
	}
}

// Register server http handlers.
func (s *Server) RegisterHandlers() {
	s.router.HandleFunc("/proxy/{uri:.*}", s.handleProxyRequest).Methods("GET", "POST", "HEAD", "PUT", "PATCH", "DELETE")
	s.router.NotFoundHandler = s.notFoundHandler()
	s.router.MethodNotAllowedHandler = s.methodNotAllowedHandler()
}

// Start proxy server.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
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

// Proxy request handler.
func (s *Server) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	logMessage := []string{fmt.Sprintf(`[%s %s] - "%s %s"`, r.RemoteAddr, r.UserAgent(), r.Method, r.URL.String())}
	// Log message should be printed only when handling is completed
	defer func(entries *[]string) {
		s.log.Println(strings.Join(*entries, " - "))
	}(&logMessage)
	// Make sure that "uri" are presents
	uri, uriFound := mux.Vars(r)["uri"]
	if !uriFound {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot extract requested path"))
		return
	}
	var schema, path = s.uriToSchemaAndPath(uri)
	if len(path) == 0 {
		s.errorHandler(w, *NewServerError(http.StatusBadRequest, "Empty request path"))
		return
	}
	var target = s.buildTargetUri(schema, path, r.URL.RawQuery)
	logMessage = append(logMessage, fmt.Sprintf("(%s)", target))
	// Create HTTP request
	hr, reqErr := http.NewRequest(r.Method, target, r.Body)
	if reqErr != nil {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot prepare http request: "+reqErr.Error()))
		return
	}
	// Proxy all request headers
	hr.Header = r.Header
	// Make an http request
	resp, respErr := s.client.Do(hr)
	// Check for response error
	if respErr != nil {
		logMessage = append(logMessage, fmt.Sprintf(`ERROR "%s"`, respErr.Error()))
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
	defer func(resp *http.Response) {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}(resp)
	logMessage = append(logMessage, fmt.Sprintf(`"HTTP %d"`, resp.StatusCode))
	// Write request response to the server response
	if writeErr := s.httpResponseToServerResponse(resp, w, true); writeErr != nil {
		s.errorHandler(w, *NewServerError(http.StatusInternalServerError, "Cannot write response:"+writeErr.Error()))
		return
	}
}

// Write HTTP request response to the server HTTP response.
func (Server) httpResponseToServerResponse(resp *http.Response, w http.ResponseWriter, addCors bool) error {
	// Read response content into buffer
	buf, err := ioutil.ReadAll(resp.Body)
	// Check for reading error
	if err != nil {
		return err
	}
	// Write headers
	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ";"))
	}
	if addCors == true {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	// Write response code
	w.WriteHeader(resp.StatusCode)
	// Write response body
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
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
func (Server) buildTargetUri(schema, domainAndPath, params string) string {
	var buf = bytes.Buffer{}
	// Write schema
	if len(schema) != 0 {
		buf.WriteString(schema)
	} else {
		buf.WriteString("http")
	}
	// Write domain and path
	buf.WriteString("://" + domainAndPath)
	// Write query params
	if len(params) != 0 {
		buf.WriteString("?" + params)
	}
	return buf.String()
}

// Schema validator.
func (Server) validateHttpSchema(schema string) bool {
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
