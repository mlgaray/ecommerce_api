package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const unknownRequestID = "unknown"

func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return unknownRequestID
	}
	return hex.EncodeToString(bytes)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := generateRequestID()

		loggerEntry := logs.WithFields(map[string]interface{}{
			"request_id":  requestID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		})

		ctx := logs.SetLogger(r.Context(), loggerEntry)
		r = r.WithContext(ctx)

		loggerEntry.WithField("event", "request_started").Info("HTTP request started")

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		loggerEntry.WithFields(logrus.Fields{
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
			"event":       "request_completed",
		}).Info("HTTP request completed")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
