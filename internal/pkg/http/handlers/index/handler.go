package index

import (
	"net/http"
)

func NewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("index")) // TODO write static HTML content
	}
}
