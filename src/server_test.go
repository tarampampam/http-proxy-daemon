package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestServer_pingHandler(t *testing.T) {
	t.Parallel()

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
		req, _  = http.NewRequest("GET", "/ping", nil)
		rr      = httptest.NewRecorder()
	)

	http.HandlerFunc(s.pingHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if expected := `"pong"`; strings.Trim(rr.Body.String(), "\n") != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServer_indexHandler(t *testing.T) {
	t.Parallel()

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "foo", testLog, testLog)
		req, _  = http.NewRequest("GET", "/", nil)
		rr      = httptest.NewRecorder()
		routes  []string
	)

	s.RegisterHandlers()

	_ = s.router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		t, err := route.GetPathTemplate()
		routes = append(routes, t)
		return err
	})

	http.HandlerFunc(s.indexHandler).ServeHTTP(rr, req)

	data := make([]string, 0)
	if err := json.Unmarshal([]byte(rr.Body.String()), &data); err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !reflect.DeepEqual(data, routes) {
		t.Errorf("handler returns wrong response: got %v want %v", data, routes)
	}
}

func TestServer_metricsHandler(t *testing.T) {
	t.Parallel()

	var (
		testLog                  = log.New(ioutil.Discard, "", 0)
		s                        = NewServer("", 8080, "", testLog, testLog)
		req, _                   = http.NewRequest("GET", "/metrics", nil)
		rr                       = httptest.NewRecorder()
		wantCode                 = http.StatusOK
		wantHostname, _          = os.Hostname()
		wantProxiedErrors  int64 = 555
		wantProxiedSuccess int64 = 666
		wantUptimeSec            = 0
		wantVersion              = VERSION
	)

	s.counters.Set(metricProxiedErrors, wantProxiedErrors)
	s.counters.Set(metricProxiedSuccess, wantProxiedSuccess)
	s.startTime = time.Now()

	http.HandlerFunc(s.metricsHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != wantCode {
		t.Errorf("handler returned wrong status code: got %v want %v", status, wantCode)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(rr.Body.String()), &data); err != nil {
		t.Fatal(err)
	}

	if hostname, _ := data["hostname"].(string); hostname != wantHostname {
		t.Errorf("unexpected hostname: got %v want %v", hostname, wantHostname)
	}

	if count := int64(math.Round(data["proxied_errors"].(float64))); count != wantProxiedErrors {
		t.Errorf("unexpected errors count: got %v want %v", count, wantProxiedErrors)
	}

	if count := int64(math.Round(data["proxied_success"].(float64))); count != wantProxiedSuccess {
		t.Errorf("unexpected successes count: got %v want %v", count, wantProxiedSuccess)
	}

	if count := int(math.Round(data["uptime_sec"].(float64))); count != wantUptimeSec {
		t.Errorf("unexpected uptime: got %v want %v", count, wantUptimeSec)
	}

	if ver, _ := data["version"].(string); ver != wantVersion {
		t.Errorf("unexpected version: got %v want %v", ver, wantVersion)
	}
}

func TestServer_notFoundHandler(t *testing.T) {
	t.Parallel()

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
		req, _  = http.NewRequest("GET", "/random_string_should_be_here_404", nil)
		rr      = httptest.NewRecorder()
	)

	s.notFoundHandler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(rr.Body.String()), &data); err != nil {
		t.Fatal(err)
	}

	if code := math.Round(data["code"].(float64)); code != http.StatusNotFound {
		t.Errorf("unexpected code in response content: got %v want %v", code, http.StatusNotFound)
	}

	if isError, _ := data["error"].(bool); isError != true {
		t.Errorf("unexpected error value in response: got %v want %v", isError, true)
	}

	if msg, _ := data["message"].(string); msg != "Not found" {
		t.Errorf("unexpected message in response: got %v want %v", msg, "Not found")
	}
}

func TestServer_methodNotAllowedHandler(t *testing.T) {
	t.Parallel()

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
		req, _  = http.NewRequest("DELETE", "/ping", nil)
		rr      = httptest.NewRecorder()
	)

	s.methodNotAllowedHandler().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(rr.Body.String()), &data); err != nil {
		t.Fatal(err)
	}

	if code := math.Round(data["code"].(float64)); code != http.StatusMethodNotAllowed {
		t.Errorf("unexpected code in response content: got %v want %v", code, http.StatusMethodNotAllowed)
	}

	if isError, _ := data["error"].(bool); isError != true {
		t.Errorf("unexpected error value in response: got %v want %v", isError, true)
	}

	if msg, _ := data["message"].(string); msg != "Method not allowed" {
		t.Errorf("unexpected message in response: got %v want %v", msg, "Method not allowed")
	}
}

func TestServer_proxyRequestHandler(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		name        string
		method      string
		uri         string
		wantCode    int
		wantStrings []string
	}{
		{
			name:        "Test GET method",
			method:      "GET",
			uri:         "https/httpbin.org/get",
			wantCode:    http.StatusOK,
			wantStrings: []string{"https://httpbin.org/get", "args", "origin", "headers"},
		},
		{
			name:        "Test POST method",
			method:      "POST",
			uri:         "https/httpbin.org/post",
			wantCode:    http.StatusOK,
			wantStrings: []string{"https://httpbin.org/post", "args", "origin", "files", "form", "headers"},
		},
		{
			name:        "Test HEAD method",
			method:      "HEAD",
			uri:         "https/httpbin.org/get",
			wantCode:    http.StatusOK,
			wantStrings: []string{},
		},
		{
			name:        "Test PUT method",
			method:      "PUT",
			uri:         "https/httpbin.org/put",
			wantCode:    http.StatusOK,
			wantStrings: []string{"https://httpbin.org/put", "args", "origin", "files", "form", "headers"},
		},
		{
			name:        "Test PATCH method",
			method:      "PATCH",
			uri:         "https/httpbin.org/patch",
			wantCode:    http.StatusOK,
			wantStrings: []string{"https://httpbin.org/patch", "args", "origin", "files", "form", "headers"},
		},
		{
			name:        "Test DELETE method",
			method:      "DELETE",
			uri:         "https/httpbin.org/delete",
			wantCode:    http.StatusOK,
			wantStrings: []string{"https://httpbin.org/delete", "args", "origin", "files", "form", "headers"},
		},
		{
			name:        "Test GET on 404 status code",
			method:      "GET",
			uri:         "https/httpbin.org/status/404",
			wantCode:    http.StatusNotFound,
			wantStrings: []string{},
		},
		{
			name:        "Test POST on 404 status code",
			method:      "POST",
			uri:         "https/httpbin.org/status/404",
			wantCode:    http.StatusNotFound,
			wantStrings: []string{},
		},
		{
			name:        "Test GET on 500 status code",
			method:      "GET",
			uri:         "https/httpbin.org/status/500",
			wantCode:    http.StatusInternalServerError,
			wantStrings: []string{},
		},
		{
			name:        "Test POST on 500 status code",
			method:      "POST",
			uri:         "https/httpbin.org/status/500",
			wantCode:    http.StatusInternalServerError,
			wantStrings: []string{},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				testLog = log.New(ioutil.Discard, "", 0)
				s       = NewServer("", 8080, "", testLog, testLog)
				req, _  = http.NewRequest(testCase.method, testCase.uri, nil)
				rr      = httptest.NewRecorder()
			)

			req = mux.SetURLVars(req, map[string]string{"uri": testCase.uri})
			http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

			if status := rr.Code; status != testCase.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, testCase.wantCode)
			}

			for _, substring := range testCase.wantStrings {
				if !strings.Contains(rr.Body.String(), substring) {
					t.Errorf("not found expected substring [%v] in response [%v]", substring, rr.Body.String())
				}
			}
		})
	}
}

func TestServer_proxyRequestHandlerErrors(t *testing.T) {
	t.Parallel()

	type BeforeExecution func(s *Server, r *http.Request) (*Server, *http.Request)

	var cases = []struct {
		name        string
		method      string
		uri         string
		runBefore   BeforeExecution
		wantCode    int
		wantContent string
	}{
		{
			name:        "Without URI (mux) inited",
			method:      "GET",
			runBefore:   nil,
			wantCode:    http.StatusInternalServerError,
			wantContent: "Cannot extract requested path",
		},
		{
			name:   "With empty request path",
			method: "GET",
			runBefore: func(s *Server, req *http.Request) (*Server, *http.Request) {
				req = mux.SetURLVars(req, map[string]string{"uri": ""})
				return s, req
			},
			wantCode:    http.StatusBadRequest,
			wantContent: "Empty request path",
		},
		{
			name:   "With response timeout",
			method: "GET",
			runBefore: func(s *Server, req *http.Request) (*Server, *http.Request) {
				req = mux.SetURLVars(req, map[string]string{"uri": "https/httpbin.org/delay/2"})
				s.SetClientResponseTimeout(time.Millisecond * 50)
				return s, req
			},
			wantCode:    http.StatusRequestTimeout,
			wantContent: "Request timeout exceeded",
		},
		{
			name:   "With wrong hostname",
			method: "GET",
			runBefore: func(s *Server, req *http.Request) (*Server, *http.Request) {
				req = mux.SetURLVars(req, map[string]string{"uri": "https/foo.invalid"})
				return s, req
			},
			wantCode:    http.StatusServiceUnavailable,
			wantContent: "dial tcp: lookup foo.invalid",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				testLog                      = log.New(ioutil.Discard, "", 0)
				s                            = NewServer("", 8080, "", testLog, testLog)
				req, _                       = http.NewRequest("GET", "", nil)
				rr                           = httptest.NewRecorder()
				wantContentTypeHeader        = "application/json"
				wantContentTypeOptionsHeader = "nosniff"
			)

			if testCase.runBefore != nil {
				s, req = testCase.runBefore(s, req)
			}

			http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

			if value := rr.Header().Get("Content-Type"); value != wantContentTypeHeader {
				t.Errorf(
					"Response has wrong Content-Type header: got %v, want %v", value, wantContentTypeHeader,
				)
			}

			if value := rr.Header().Get("X-Content-Type-Options"); value != wantContentTypeOptionsHeader {
				t.Errorf(
					"Response has wrong X-Content-Type-Options header: got %v, want %v", value, wantContentTypeOptionsHeader,
				)
			}

			if status := rr.Code; status != testCase.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, testCase.wantCode)
			}

			if !strings.Contains(rr.Body.String(), testCase.wantContent) {
				t.Errorf("not found expected substring [%v] in response [%v]", testCase.wantContent, rr.Body.String())
			}
		})
	}
}

func TestServer_RegisterHandlers(t *testing.T) {
	t.Parallel()

	compareHandlers := func(h1, h2 interface{}) bool {
		t.Helper()
		return reflect.ValueOf(h1).Pointer() == reflect.ValueOf(h2).Pointer()
	}

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "proxy", testLog, testLog)
	)

	var cases = []struct {
		name    string
		route   string
		methods []string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{
			name:    "index",
			route:   "/",
			methods: []string{"GET"},
			handler: s.indexHandler,
		},
		{
			name:    "ping",
			route:   "/ping",
			methods: []string{"GET"},
			handler: s.pingHandler,
		},
		{
			name:    "metrics",
			route:   "/metrics",
			methods: []string{"GET"},
			handler: s.metricsHandler,
		},
		{
			name:    "proxy",
			route:   "/proxy/{uri:.*}",
			methods: []string{"GET", "POST", "HEAD", "PUT", "PATCH", "DELETE", "OPTIONS"},
			handler: s.proxyRequestHandler,
		},
	}

	for _, testCase := range cases {
		if s.router.Get(testCase.name) != nil {
			t.Errorf("Handler for route [%s] must be not registered before RegisterHandlers() calling", testCase.name)
		}
	}

	s.RegisterHandlers()

	for _, testCase := range cases {
		if route, _ := s.router.Get(testCase.name).GetPathTemplate(); route != testCase.route {
			t.Errorf("wrong route for [%s] route: want %v, got %v", testCase.name, testCase.route, route)
		}
		if methods, _ := s.router.Get(testCase.name).GetMethods(); !reflect.DeepEqual(methods, testCase.methods) {
			t.Errorf("wrong method(s) for [%s] route: want %v, got %v", testCase.name, testCase.methods, methods)
		}
		if !compareHandlers(testCase.handler, s.router.Get(testCase.name).GetHandler()) {
			t.Errorf("wrong handler for [%s] route", testCase.name)
		}
	}

	if !compareHandlers(s.router.NotFoundHandler, s.notFoundHandler()) {
		t.Error("Wrong NotFound handler")
	}

	if !compareHandlers(s.router.MethodNotAllowedHandler, s.methodNotAllowedHandler()) {
		t.Error("Wrong NotFound handler")
	}
}

func TestServer_SetClientResponseTimeout(t *testing.T) {
	t.Parallel()

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "proxy", testLog, testLog)
	)

	if s.client.Timeout != time.Second*30 {
		t.Error("Unexpected default timeout")
	}

	s.SetClientResponseTimeout(time.Hour)

	if s.client.Timeout != time.Hour {
		t.Error("Unexpected new timeout")
	}
}

func TestServer_validateHttpSchema(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		give string
		want bool
	}{
		{
			give: "http",
			want: true,
		},
		{
			give: "https",
			want: true,
		},
		{
			give: "hTTpS",
			want: false,
		},
		{
			give: "hTTp",
			want: false,
		},
		{
			give: "foo",
			want: false,
		},
		{
			give: "",
			want: false,
		},
		{
			give: "foo bar baz",
			want: false,
		},
	}

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
	)

	for _, testCase := range cases {
		if s.validateHttpSchema(testCase.give) != testCase.want {
			t.Errorf("For [%s] must returns %+v", testCase.give, testCase.want)
		}
	}
}

func TestServer_buildTargetUri(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		schema, domainAndPath, params string
		want                          string
	}{
		{
			schema:        "https",
			domainAndPath: "google.com",
			params:        "foo=bar",
			want:          "https://google.com?foo=bar",
		},
		{
			schema:        "https",
			domainAndPath: "google.com/some/shit",
			params:        "foo=bar&bar=baz",
			want:          "https://google.com/some/shit?foo=bar&bar=baz",
		},
		{
			schema:        "",
			domainAndPath: "google.com",
			params:        "",
			want:          "http://google.com",
		},
		{
			schema:        "ftp",
			domainAndPath: "google.com",
			params:        "",
			want:          "ftp://google.com",
		},
		{
			schema:        "",
			domainAndPath: "",
			params:        "",
			want:          "",
		},
		{
			schema:        "",
			domainAndPath: "a",
			params:        "",
			want:          "http://a",
		},
	}

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
	)

	for _, testCase := range cases {
		if s.buildTargetUri(testCase.schema, testCase.domainAndPath, testCase.params) != testCase.want {
			t.Errorf(
				"For [%s, %s, %s] must returns %s",
				testCase.schema,
				testCase.domainAndPath,
				testCase.params,
				testCase.want,
			)
		}
	}
}

func TestServer_uriToSchemaAndPath(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		uri, wantSchema, wantPath string
	}{
		{
			uri:        "https/google.com",
			wantSchema: "https",
			wantPath:   "google.com",
		},
		{
			uri:        "http/google.com",
			wantSchema: "http",
			wantPath:   "google.com",
		},
		{
			uri:        "hTTps/google.COM",
			wantSchema: "https",
			wantPath:   "google.COM",
		},
		{
			uri:        "google.com",
			wantSchema: "",
			wantPath:   "google.com",
		},
		{
			uri:        "google.com/foo?bar=baz",
			wantSchema: "",
			wantPath:   "google.com/foo?bar=baz",
		},
		{
			uri:        "ftp/google.com/foo?bar=baz",
			wantSchema: "",
			wantPath:   "ftp/google.com/foo?bar=baz",
		},
	}

	var (
		testLog = log.New(ioutil.Discard, "", 0)
		s       = NewServer("", 8080, "", testLog, testLog)
	)

	for _, testCase := range cases {
		gotSchema, gotPath := s.uriToSchemaAndPath(testCase.uri)
		if gotSchema != testCase.wantSchema || gotPath != testCase.wantPath {
			t.Errorf(
				"For [%s] must returns schema [%s] and path [%s], but returns [%s, %s]",
				testCase.uri,
				testCase.wantSchema,
				testCase.wantPath,
				gotSchema, gotPath,
			)
		}
	}
}
