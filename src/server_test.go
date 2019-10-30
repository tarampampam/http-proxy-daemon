package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
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

func TestServer_proxyRequestHandlerWithoutUriInit(t *testing.T) {
	t.Parallel()

	var (
		testLog     = log.New(ioutil.Discard, "", 0)
		s           = NewServer("", 8080, "", testLog, testLog)
		req, _      = http.NewRequest("GET", "https/httpbin.org/anything", nil)
		rr          = httptest.NewRecorder()
		wantCode    = http.StatusInternalServerError
		wantContent = "Cannot extract requested path"
	)

	http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != wantCode {
		t.Errorf("handler returned wrong status code: got %v want %v", status, wantCode)
	}

	if !strings.Contains(rr.Body.String(), wantContent) {
		t.Errorf("not found expected substring [%v] in response [%v]", wantContent, rr.Body.String())
	}
}

func TestServer_proxyRequestHandlerWithEmptyRequestPath(t *testing.T) {
	t.Parallel()

	var (
		testLog     = log.New(ioutil.Discard, "", 0)
		s           = NewServer("", 8080, "", testLog, testLog)
		req, _      = http.NewRequest("GET", "", nil)
		rr          = httptest.NewRecorder()
		wantCode    = http.StatusBadRequest
		wantContent = "Empty request path"
	)

	req = mux.SetURLVars(req, map[string]string{"uri": ""})
	http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != wantCode {
		t.Errorf("handler returned wrong status code: got %v want %v", status, wantCode)
	}

	if !strings.Contains(rr.Body.String(), wantContent) {
		t.Errorf("not found expected substring [%v] in response [%v]", wantContent, rr.Body.String())
	}
}

func TestServer_proxyRequestHandlerWithResponseTimeout(t *testing.T) {
	t.Parallel()

	var (
		testLog     = log.New(ioutil.Discard, "", 0)
		s           = NewServer("", 8080, "", testLog, testLog)
		req, _      = http.NewRequest("GET", "https/httpbin.org/delay/2", nil)
		rr          = httptest.NewRecorder()
		wantCode    = http.StatusRequestTimeout
		wantContent = "Request timeout exceeded"
	)

	s.SetClientResponseTimeout(time.Millisecond * 50)

	req = mux.SetURLVars(req, map[string]string{"uri": "https/httpbin.org/delay/2"})
	http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != wantCode {
		t.Errorf("handler returned wrong status code: got %v want %v", status, wantCode)
	}

	if !strings.Contains(rr.Body.String(), wantContent) {
		t.Errorf("not found expected substring [%v] in response [%v]", wantContent, rr.Body.String())
	}
}

func TestServer_proxyRequestHandlerWithWrongHostname(t *testing.T) {
	t.Parallel()

	var (
		testLog     = log.New(ioutil.Discard, "", 0)
		s           = NewServer("", 8080, "", testLog, testLog)
		req, _      = http.NewRequest("GET", "https/foo.invalid", nil)
		rr          = httptest.NewRecorder()
		wantCode    = http.StatusServiceUnavailable
		wantContent = "no such host"
	)

	req = mux.SetURLVars(req, map[string]string{"uri": "https/foo.invalid"})
	http.HandlerFunc(s.proxyRequestHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != wantCode {
		t.Errorf("handler returned wrong status code: got %v want %v", status, wantCode)
	}

	if !strings.Contains(rr.Body.String(), wantContent) {
		t.Errorf("not found expected substring [%v] in response [%v]", wantContent, rr.Body.String())
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
