package http

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "ecommerce-api",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}
