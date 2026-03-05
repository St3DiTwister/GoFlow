package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProcessedEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goflow_worker_processed_events_total",
		Help: "Total number of events processed by worker",
	})

	InvalidEvents = promauto.NewCounter(prometheus.CounterOpts{
		Name: "goflow_worker_invalid_events_total",
		Help: "Total number of events with invalid site_id",
	})

	ClickHouseInsertDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "goflow_worker_ch_insert_duration_seconds",
		Help:    "Duration of ClickHouse insert operations",
		Buckets: []float64{0.1, 0.5, 1, 2, 5},
	})
)
