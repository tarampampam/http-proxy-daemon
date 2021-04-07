package index

import (
	"net/http"
)

func NewHandler() http.HandlerFunc {
	content := []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow" />
    <title>HTTP Proxy Daemon</title>
    <link rel="dns-prefetch" href="//fonts.gstatic.com">
    <link href="https://fonts.googleapis.com/css?family=Nunito" rel="stylesheet">
    <style>
        html,body {background-color:#2f2f2f;color:#fff;font-family:'Nunito',sans-serif;
                   font-weight:100;height:100vh;margin:0}
        .full-height {height:100vh}
        .flex-center {align-items:center;display:flex;justify-content:center}
        .position-ref {position:relative}
        .message {font-size:26px;padding:0 0 0 15px}
    </style>
</head>
<body>
<div class="flex-center position-ref full-height">
    <img src="https://hsto.org/webt/gd/ek/am/gdekam4xddqshipd99rku40fec4.png" width="60" height="60" alt=""/>
    <div class="message">
        HTTP Proxy Daemon
    </div>
</div>
</body>
</html>`)

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(content)
	}
}
