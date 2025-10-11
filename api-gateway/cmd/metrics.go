package main

import "github.com/prometheus/client_golang/prometheus"

var (
	createShortURLTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "create_short_url_total",
			Help: "Total number of short URL creation requests",
		},
		[]string{"status"}, // Labels: success, error, invalid_url, invalid_expiration
	)

	getLongURLTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "get_long_url_total",
			Help: "Total number of long URL retrieval requests",
		},
		[]string{"status"}, // Labels: success, error, not_found, expired
	)

	deleteShortURLTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "delete_short_url_total",
			Help: "Total number of short URL deletion requests",
		},
		[]string{"status"}, // Labels: success, error, not_found
	)

	getAllURLsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "get_all_urls_total",
			Help: "Total number of get all URLs requests",
		},
		[]string{"status"}, // Labels: success, error, no_content
	)
)

func initMetrics() {
	// Register Prometheus metrics
	prometheus.MustRegister(createShortURLTotal)
	prometheus.MustRegister(getLongURLTotal)
	prometheus.MustRegister(deleteShortURLTotal)
	prometheus.MustRegister(getAllURLsTotal)
}
