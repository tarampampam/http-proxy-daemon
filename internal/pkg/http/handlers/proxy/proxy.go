package proxy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type metrics interface {
	IncrementSuccessful()
	IncrementFailed()
	IncrementErrors()
}

type Handler struct {
	ctx        context.Context
	httpClient httpClient
	m          metrics
}

const (
	proxyErrPrefix      = "proxy: "
	defaultTargetSchema = "http"
)

func NewHandler(ctx context.Context, httpClient httpClient, m metrics) *Handler {
	return &Handler{ctx: ctx, httpClient: httpClient, m: m}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nolint:funlen
	// make sure that "uri" are presents
	uri, uriFound := mux.Vars(r)["uri"]
	if !uriFound {
		h.m.IncrementErrors()
		http.Error(w, proxyErrPrefix+"cannot extract requested URI", http.StatusInternalServerError)

		return
	}

	// extract request schema and path from requested uri
	var schema, path = h.uriToSchemaAndPath(uri) // schema is optional
	if path == "" {
		h.m.IncrementErrors()
		http.Error(w, proxyErrPrefix+"empty request path", http.StatusBadRequest)

		return
	}

	// build target uri
	targetURI, targetURIErr := h.buildTargetURI(schema, path, r.URL.RawQuery)
	if targetURIErr != nil {
		h.m.IncrementErrors()
		http.Error(w, proxyErrPrefix+"cannot build target URI", http.StatusBadRequest)

		return
	}

	// create an HTTP request
	req, reqErr := http.NewRequestWithContext(h.ctx, r.Method, targetURI, r.Body)
	if reqErr != nil {
		h.m.IncrementErrors()
		http.Error(w, proxyErrPrefix+reqErr.Error(), http.StatusInternalServerError)

		return
	}

	// proxy all request headers
	req.Header = r.Header.Clone()

	// make an http request
	resp, respErr := h.httpClient.Do(req)
	if respErr != nil {
		defer h.m.IncrementFailed()

		if e, ok := respErr.(*url.Error); ok && e.Timeout() { //nolint:errorlint
			http.Error(w, proxyErrPrefix+"request timeout exceeded", http.StatusRequestTimeout)

			return
		}

		http.Error(w, proxyErrPrefix+respErr.Error(), http.StatusServiceUnavailable)

		return
	}

	defer func() { _ = resp.Body.Close() }()

	// write HTTP response headers into current HTTP request headers
	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ";"))
	}

	// allow access from anywhere
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(resp.StatusCode)

	if _, copyErr := io.Copy(w, resp.Body); copyErr != nil {
		h.m.IncrementErrors()
		http.Error(w, proxyErrPrefix+copyErr.Error(), http.StatusInternalServerError)

		return
	}

	h.m.IncrementSuccessful()
}

func (h *Handler) uriToSchemaAndPath(uri string) (string, string) {
	slashPos := strings.IndexByte(uri, '/')

	if slashPos != -1 {
		schema := strings.ToLower(uri[:slashPos])

		if (schema == "http" || schema == "https") && len(uri) > slashPos+1 {
			return schema, uri[slashPos+1:]
		}
	}

	return "", uri
}

func (h *Handler) buildTargetURI(schema, path, params string) (string, error) {
	var b strings.Builder

	b.Grow(len(schema) + len(path) + len(params) + 3) //nolint:gomnd

	if len(schema) != 0 {
		b.WriteString(schema)
	} else {
		b.WriteString(defaultTargetSchema)
	}

	b.WriteString("://" + path)

	if params != "" {
		b.WriteString("?" + params)
	}

	if b.Len() < 10 { //nolint:gomnd
		return "", errors.New("target URI building error")
	}

	return b.String(), nil
}
