package middleware

import (
	"net/http"
	"time"
)

// Logger interface for request logging
type Logger interface {
	Info(msg string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// RequestLoggingMiddleware creates a middleware that logs HTTP requests
func RequestLoggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(lrw, r)

			// Log request
			duration := time.Since(start)
			logger.Info("HTTP request",
				Field{Key: "method", Value: r.Method},
				Field{Key: "path", Value: r.URL.Path},
				Field{Key: "status", Value: lrw.statusCode},
				Field{Key: "duration", Value: duration.String()},
				Field{Key: "remote_addr", Value: r.RemoteAddr},
			)
		})
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}
