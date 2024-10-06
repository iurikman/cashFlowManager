package rest

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	requestsDuration *prometheus.HistogramVec
}

const (
	namespace = "request_duration"
	subsystem = "wallet_service"
)

func newMetrics() *metrics {
	return &metrics{
		promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_duration",
			Help:      "transaction_duration",
		},
			[]string{"method", "path"},
		),
	}
}
