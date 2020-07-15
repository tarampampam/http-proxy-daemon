package proxy

import (
	"crypto/tls"
	"errors"
	"http-proxy-daemon/counters"
	serverErrors "http-proxy-daemon/http/errors"
	"http-proxy-daemon/shared"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	httpClient *http.Client
	counters   counters.Counters
}

func NewHandler(counters counters.Counters, httpRequestTimeout time.Duration, maxRedirects int) http.Handler {
	return &Handler{
		counters: counters,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec //lgtm [go/disabled-certificate-check]
				},
			},
			Timeout: httpRequestTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return errors.New("too many (" + strconv.Itoa(maxRedirects) + ") redirects")
				}
				return nil
			},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nolint:funlen
	var success bool = false

	defer func(success *bool) {
		var metricName string

		if success != nil && *success {
			metricName = shared.MetricProxiedSuccess
		} else {
			metricName = shared.MetricProxiedErrors
		}

		h.counters.Increment(metricName)
	}(&success)

	// make sure that "uri" are presents
	uri, uriFound := mux.Vars(r)["uri"]
	if !uriFound {
		h.respondWithError(w, http.StatusInternalServerError, "Cannot extract requested path")

		return
	}

	// extract request schema and path from route
	var schema, path = h.uriToSchemaAndPath(uri)
	if len(path) == 0 {
		h.respondWithError(w, http.StatusBadRequest, "Empty request path")

		return
	}

	// create an HTTP request
	httpRequest, reqErr := http.NewRequest(r.Method, h.buildTargetURI(schema, path, r.URL.RawQuery), nil)
	if reqErr != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Cannot prepare http request: "+reqErr.Error())

		return
	}

	// proxy all request headers
	httpRequest.Header = r.Header.Clone()

	// Make an http request
	response, responseErr := h.httpClient.Do(httpRequest)

	if responseErr != nil {
		if e, ok := responseErr.(*url.Error); ok {
			if e.Timeout() {
				h.respondWithError(w, http.StatusRequestTimeout, "Request timeout exceeded")

				return
			}
		}

		h.respondWithError(w, http.StatusServiceUnavailable, responseErr.Error())

		return
	}

	defer response.Body.Close()

	h.setHeaders(w, response)

	if _, copyErr := io.Copy(w, response.Body); copyErr != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Cannot write response: "+copyErr.Error())

		return
	}

	success = true
}

func (h *Handler) setHeaders(w http.ResponseWriter, httpResponse *http.Response) {
	// write HTTP response headers into current HTTP request headers
	for k, v := range httpResponse.Header {
		w.Header().Set(k, strings.Join(v, ";"))
	}

	// allow access from anywhere
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// write response code
	w.WriteHeader(httpResponse.StatusCode)
}

func (h *Handler) respondWithError(w http.ResponseWriter, code int, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(serverErrors.NewServerError(code, error).ToJSON())
}

// Extract schema and path from passed specific uri string.
func (h *Handler) uriToSchemaAndPath(uri string) (schema string, path string) {
	if strings.Contains(uri, "/") {
		slashPos := strings.IndexByte(uri, '/')
		possibleSchema := strings.ToLower(uri[:slashPos])

		if possibleSchema == "http" || possibleSchema == "https" {
			schema, path = possibleSchema, uri[slashPos+1:]
		}
	}

	return
}

func (h *Handler) buildTargetURI(schema, path, params string) (uri string) {
	if len(schema) != 0 {
		uri += schema
	} else {
		uri += "http"
	}

	uri += "://" + path

	if len(params) != 0 {
		uri += "?" + params
	}

	if len(uri) < 8 { //nolint:gomnd
		return ""
	}

	return
}
