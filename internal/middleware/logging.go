package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/your-username/maintenance-api/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Add request ID to context
			ctx := logger.WithRequestID(r.Context())
			r = r.WithContext(ctx)
			
			// Create response writer with status tracking
			lrw := newLoggingResponseWriter(w)
			
			// Process request
			next.ServeHTTP(lrw, r)
			
			// Log request details
			duration := time.Since(start)
			log.WithContext(ctx).LogRequest(
				r.Method,
				r.URL.Path,
				lrw.statusCode,
				duration,
				nil,
			)
		})
	}
}

// RequestLogger is a convenience function for logging specific requests
func RequestLogger(log *logger.Logger, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithRequestID(r.Context())
		r = r.WithContext(ctx)
		
		start := time.Now()
		handler(w, r)
		duration := time.Since(start)
		
		log.WithContext(ctx).LogRequest(
			r.Method,
			r.URL.Path,
			http.StatusOK, // Default, can be overridden by handler
			duration,
			nil,
		)
	}
}