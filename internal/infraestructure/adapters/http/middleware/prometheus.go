package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Request metrics
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

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{1, 10, 100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{1, 10, 100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Concurrent requests
	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed",
	})

	// Error rate by status code family
	httpRequestsByStatusFamily = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_by_status_family_total",
			Help: "Total HTTP requests by status code family (2xx, 3xx, 4xx, 5xx)",
		},
		[]string{"status_family", "endpoint"},
	)
)

type prometheusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (prw *prometheusResponseWriter) WriteHeader(code int) {
	prw.statusCode = code
	prw.ResponseWriter.WriteHeader(code)
}

func (prw *prometheusResponseWriter) Write(b []byte) (int, error) {
	size, err := prw.ResponseWriter.Write(b)
	prw.size += size
	return size, err
}

func getStatusFamily(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func getEndpoint(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if template, err := route.GetPathTemplate(); err == nil {
			return template
		}
	}
	return r.URL.Path
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment in-flight requests
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Wrap response writer to capture status code and response size
		prw := &prometheusResponseWriter{
			ResponseWriter: w,
			statusCode:     200, // default status code
			size:          0,
		}

		// Record request size
		requestSize := float64(r.ContentLength)
		if requestSize < 0 {
			requestSize = 0
		}

		endpoint := getEndpoint(r)
		method := r.Method

		httpRequestSize.WithLabelValues(method, endpoint).Observe(requestSize)

		// Process request
		next.ServeHTTP(prw, r)

		// Calculate duration
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(prw.statusCode)
		statusFamily := getStatusFamily(prw.statusCode)

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
		httpResponseSize.WithLabelValues(method, endpoint, statusCode).Observe(float64(prw.size))
		httpRequestsByStatusFamily.WithLabelValues(statusFamily, endpoint).Inc()
	})
}