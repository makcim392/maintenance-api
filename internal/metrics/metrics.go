package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Application metrics
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	tasksCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_created_total",
			Help: "Total number of tasks created",
		},
	)

	tasksUpdated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_updated_total",
			Help: "Total number of tasks updated",
		},
	)

	tasksDeleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tasks_deleted_total",
			Help: "Total number of tasks deleted",
		},
	)

	// Database metrics
	dbConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Authentication metrics
	authAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total authentication attempts",
		},
		[]string{"method", "status"},
	)

	// Error metrics
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "operation"},
	)
)

// MetricsMiddleware wraps HTTP handlers to collect metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Increment active connections
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Create a response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process the request
		next.ServeHTTP(rw, r)
		
		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(rw.statusCode)
		
		httpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			statusCode,
		).Inc()
		
		httpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			statusCode,
		).Observe(duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecordTaskCreated increments the tasks created counter
func RecordTaskCreated() {
	tasksCreated.Inc()
}

// RecordTaskUpdated increments the tasks updated counter
func RecordTaskUpdated() {
	tasksUpdated.Inc()
}

// RecordTaskDeleted increments the tasks deleted counter
func RecordTaskDeleted() {
	tasksDeleted.Inc()
}

// RecordAuthAttempt records authentication attempts
func RecordAuthAttempt(method string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	authAttempts.WithLabelValues(method, status).Inc()
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordError records error metrics
func RecordError(errorType, operation string) {
	errorsTotal.WithLabelValues(errorType, operation).Inc()
}

// SetActiveDBConnections sets the active database connections gauge
func SetActiveDBConnections(count int) {
	dbConnections.Set(float64(count))
}

// MetricsHandler returns the Prometheus metrics handler
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// ContextKey for metrics
type contextKey string

const (
	RequestStartTimeKey contextKey = "request_start_time"
)

// WithRequestStartTime adds the request start time to context
func WithRequestStartTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestStartTimeKey, time.Now())
}

// GetRequestDuration returns the duration since request started
func GetRequestDuration(ctx context.Context) time.Duration {
	if startTime, ok := ctx.Value(RequestStartTimeKey).(time.Time); ok {
		return time.Since(startTime)
	}
	return 0
}