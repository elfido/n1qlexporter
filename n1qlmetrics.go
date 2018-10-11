package main

import "github.com/prometheus/client_golang/prometheus"

// metrics

// Active queries
var activeExecutionTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_active_time_execution",
		Help:    "N1QL Current queries execution time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 17),
	},
	[]string{"cluster", "node", "query_type"},
)

var activeAccumulation = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_active_accumulated_queries",
		Help:    "N1QL Current queries in execution",
		Buckets: []float64{0, 10, 20, 50, 100, 250, 1000, 5000, 10000},
	},
	[]string{"cluster", "node"},
)

var activeWaitingTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_active_time_waiting",
		Help:    "N1QL Current queries waiting time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 17),
	},
	[]string{"cluster", "node", "query_type"},
)

var activeScanConsistency = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "n1ql_active_consistency",
		Help: "N1QL Current queries waiting time",
	},
	[]string{"cluster", "consistency"},
)

// Completed queries
var completedResultCount = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_completed_result_count",
		Help:    "N1QL Number of results per query",
		Buckets: []float64{0, 10, 20, 50, 100, 250, 500, 1000, 5000, 10000, 100000, 500000, 1000000},
	},
	[]string{"cluster", "query_type"},
)

var completedResultSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_completed_result_size",
		Help:    "N1QL Response size in bytes",
		Buckets: prometheus.ExponentialBuckets(200, 2.5, 15),
	},
	[]string{"cluster", "query_type"},
)

var completedExecutionTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_completed_time_execution",
		Help:    "N1QL Current queries execution time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 17),
	},
	[]string{"cluster", "node", "query_type", "state"},
)

var completedWaitingTime = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "n1ql_completed_time_waiting",
		Help:    "N1QL Completed queries waiting time",
		Buckets: prometheus.ExponentialBuckets(1, 2, 17),
	},
	[]string{"cluster", "node", "query_type"},
)

var completedPrimaryIndexUse = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "n1ql_completed_primaryindex",
		Help: "N1QL Current queries waiting time",
	},
	[]string{"cluster", "query_type"},
)

// Vitals
var completedVitals = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "n1ql_vitals_completed_queries",
		Help: "N1QL completed queries from vitals",
	},
	[]string{"cluster", "node"},
)

var cpuVitals = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "n1ql_vitals_cpu_usage",
		Help: "N1QL CPU usage for user/system",
	},
	[]string{"cluster", "node", "space"},
)

func initN1QLMetrics() {
	prometheus.MustRegister(
		activeExecutionTime,
		activeAccumulation,
		activeWaitingTime,
		activeScanConsistency,
		// Completed queries
		completedResultCount,
		completedResultSize,
		completedExecutionTime,
		completedWaitingTime,
		completedPrimaryIndexUse,
		// Vitals
		completedVitals,
		cpuVitals,
	)
}
