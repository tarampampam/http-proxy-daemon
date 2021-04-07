package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// NewHTMLErrorHandler creates error handler, thar responds with HTML-formatted error with passed status code.
func NewHTMLErrorHandler(code int) http.Handler {
	tmpl := []byte(defaultErrorTemplate.Build(code))

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(code)
		_, _ = w.Write(tmpl)
	})
}

type errorPageTemplate string

const defaultErrorTemplate errorPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow" />
    <title>{{ message }}</title>
    <link rel="dns-prefetch" href="//fonts.gstatic.com">
    <link href="https://fonts.googleapis.com/css?family=Nunito" rel="stylesheet">
    <style>
        html,body {background-color:#2f2f2f;color:#fff;font-family:'Nunito',sans-serif;
                   font-weight:100;height:100vh;margin:0}
        .full-height {height:100vh}
        .flex-center {align-items:center;display:flex;justify-content:center}
        .position-ref {position:relative}
        .code {font-size:18px;text-align:center;padding:10px}
        .message {border-right:2px solid;font-size:26px;padding:0 10px 0 15px;text-align:center}
    </style>
</head>
<body>
<div class="flex-center position-ref full-height">
    <div class="message">
        {{ message }}
    </div>
    <div class="code">
        {{ code }}
    </div>
</div>
</body>
</html>`

// Build makes registered patterns replacing.
func (t errorPageTemplate) Build(errorCode int) string {
	out := string(t)

	for k, v := range map[string]string{
		"code":    strconv.Itoa(errorCode),
		"message": http.StatusText(errorCode),
	} {
		out = strings.ReplaceAll(out, fmt.Sprintf("{{ %s }}", k), v)
	}

	return out
}
