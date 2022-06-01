package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	// Prometheus Registry to register metrics.
	prometheusRegistry = prometheus.NewRegistry()

	totalHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "my_app_http_hit_total",
			Help: "Total number of http hits.",
		},
	)
)

func metricsHandler() http.Handler {
	return promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{})
}

func init() {
	prometheusRegistry.MustRegister(totalHits)
}

func count(f func(http.ResponseWriter, *http.Request), name string) func(http.ResponseWriter, *http.Request) {
	httpHits := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: name,
			Help: "Total number of http hits.",
		},
	)

	err := prometheusRegistry.Register(httpHits)
	if err != nil {
		return f
	}

	return func(w http.ResponseWriter, r *http.Request) {
		httpHits.Inc()
		totalHits.Inc()
		f(w, r)
	}
}
