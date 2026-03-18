package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

const (
	statusHealthy   = "healthy"
	statusUnhealthy = "unhealthy"
	statusDegraded  = "degraded"
)

type HealthHandler struct {
	db *sql.DB
}

type dependencyStatus struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
}

type healthResponse struct {
	Status       string                      `json:"status"`
	Service      string                      `json:"service"`
	Timestamp    string                      `json:"timestamp"`
	Dependencies map[string]dependencyStatus `json:"dependencies"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:    statusHealthy,
		Service:   "ecommerce-api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Dependencies: map[string]dependencyStatus{
			"database": h.checkDatabase(),
		},
	}

	httpStatus := http.StatusOK

	dbDep := resp.Dependencies["database"]
	switch dbDep.Status {
	case statusUnhealthy:
		resp.Status = statusUnhealthy
		httpStatus = http.StatusServiceUnavailable
	case statusDegraded:
		resp.Status = statusDegraded
	}

	responseData, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	w.Write(responseData)
}

func (h *HealthHandler) checkDatabase() dependencyStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.db.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		return dependencyStatus{Status: statusUnhealthy, LatencyMs: latency.Milliseconds()}
	}

	// DB ping > 500ms → degraded
	if latency > 500*time.Millisecond {
		return dependencyStatus{Status: statusDegraded, LatencyMs: latency.Milliseconds()}
	}

	return dependencyStatus{Status: statusHealthy, LatencyMs: latency.Milliseconds()}
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}
