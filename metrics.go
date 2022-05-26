package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	// Prometheus Registry to register metrics.
	prometheusRegistry = prometheus.NewRegistry()
)

func metricsHandler() http.Handler {
	return promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{})
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
		f(w, r)
	}
}
