package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PgErrCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "timeslots",
		Subsystem: "pg",
		Name:      "pg_err_count",
	}, []string{"method"})
	PgDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "timeslots",
		Subsystem: "pg",
		Name:      "pg_duration",
	}, []string{"method"})
)
