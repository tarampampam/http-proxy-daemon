package metrics

import (
	"encoding/json"
	"http-proxy-daemon/counters"
	"http-proxy-daemon/shared"
	"http-proxy-daemon/version"
	"net/http"
	"os"
	"time"
)

type (
	Handler struct {
		serverStartTime *time.Time
		counters        counters.Counters
	}

	response struct {
		ProxiedSuccess int64  `json:"proxied_success"`
		ProxiedErrors  int64  `json:"proxied_errors"`
		UptimeSec      int64  `json:"uptime_sec"`
		Hostname       string `json:"hostname"`
		Version        string `json:"version"`
	}
)

func NewHandler(serverStartTime *time.Time, counters counters.Counters) http.Handler {
	return &Handler{
		serverStartTime: serverStartTime,
		counters:        counters,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	var hostname string = ""

	// Append hostname
	if h, err := os.Hostname(); err == nil {
		hostname = h
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(response{
		ProxiedSuccess: h.counters.Get(shared.MetricProxiedSuccess),
		ProxiedErrors:  h.counters.Get(shared.MetricProxiedErrors),
		UptimeSec:      int64(time.Since(*h.serverStartTime).Seconds()),
		Version:        version.Version(),
		Hostname:       hostname,
	})
}
