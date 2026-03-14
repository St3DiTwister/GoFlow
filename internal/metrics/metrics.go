package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProcessedEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goflow_worker_processed_events_total",
		Help: "Total number of events successfully sent to ClickHouse",
	}, []string{"pod"})

	RejectedEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goflow_worker_rejected_events_total",
		Help: "Total number of events rejected by worker logic",
	}, []string{"reason", "pod"})

	SystemErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goflow_worker_errors_total",
		Help: "Total number of technical errors in worker",
	}, []string{"op", "pod"})

	ClickHouseInsertDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "goflow_worker_ch_insert_duration_seconds",
		Help:    "Duration of ClickHouse insert operations",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"pod"})
)
