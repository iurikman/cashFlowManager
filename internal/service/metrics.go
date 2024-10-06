package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	xrRequests *prometheus.CounterVec
}

func newMetrics() *metrics {
	const (
		namespace = "requests"
		subsystem = "wallet_service"
	)

	return &metrics{
		promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "xr_requests_total",
			Help:      "xr_requests_total",
		},
			[]string{"currency_from", "currency_to"},
		),
	}
}

func (m *metrics) IncrXRRequests(currencyFrom string, currencyTo string) {
	m.xrRequests.WithLabelValues(currencyFrom, currencyTo).Inc()
}
