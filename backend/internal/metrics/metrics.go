package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP Metrics
var (
	// HTTPRequestsTotal counts total HTTP requests by method, endpoint, and status code
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration tracks HTTP request latency distribution
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distribution",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// HTTPRequestsInFlight tracks concurrent HTTP requests
	HTTPRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"endpoint"},
	)
)

var (
	// SMTPEmailsReceivedTotal counts total emails received via SMTP
	SMTPEmailsReceivedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "smtp_emails_received_total",
			Help: "Total number of emails received via SMTP",
		},
	)

	// SMTPEmailsRejectedTotal counts rejected emails by reason
	SMTPEmailsRejectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smtp_emails_rejected_total",
			Help: "Total number of rejected emails",
		},
		[]string{"reason"},
	)

	// SMTPSessionsActive tracks concurrent SMTP sessions
	SMTPSessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "smtp_sessions_active",
			Help: "Number of active SMTP sessions",
		},
	)
)

var (
	// ActiveAddressesTotal tracks the total number of registered addresses
	ActiveAddressesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "coresend_active_addresses_total",
			Help: "Total number of active registered email addresses",
		},
	)

	// EmailsStoredTotal tracks the total number of emails stored in the system
	EmailsStoredTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "coresend_emails_stored_total",
			Help: "Total number of emails currently stored",
		},
	)

	// EmailsPerAddress tracks the distribution of emails per address
	EmailsPerAddress = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "coresend_emails_per_address",
			Help:    "Distribution of number of emails per address",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		},
	)
)

var (
	// RateLimitHitsTotal counts rate limit hits by endpoint
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coresend_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint"},
	)

	// AuthFailuresTotal counts authentication failures by reason
	AuthFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coresend_auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"reason"},
	)
)

var (
	// RedisOperationDuration tracks latency of critical Redis operations
	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Duration of Redis operations from application",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation"},
	)
)
