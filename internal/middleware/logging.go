package middleware

import (
	"content-maestro/internal/logger"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	log := logger.NewLogger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := newLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		log.Infow("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.statusCode,
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

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
