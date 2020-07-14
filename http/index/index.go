package index

import (
	"encoding/json"
	"http-proxy-daemon/http/errors"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	router *mux.Router
}

func NewHandler(router *mux.Router) http.Handler {
	return &Handler{
		router: router,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	var routes []string

	// walk through all available routes an fill routes slice
	err := h.router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		t, err := route.GetPathTemplate()

		if err == nil {
			routes = append(routes, t)
			return nil
		}

		return err
	})

	// handle possible error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(errors.NewServerError(http.StatusInternalServerError, err.Error()).ToJSON())

		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(routes)
}
