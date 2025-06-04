package handlers

import (
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/bbcbear/scd41-exporter/internal/sensor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(sensorRef sensor.Sensor, isHealthy *atomic.Bool) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if !isHealthy.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("sensor error"))
			slog.Warn("Health check failed", "remote", r.RemoteAddr)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}
