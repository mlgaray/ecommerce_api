package ports

import "net/http"

// CategoryHandler defines the HTTP handler interface for category endpoints.
type CategoryHandler interface {
	// Create handles POST /categories requests.
	Create(w http.ResponseWriter, r *http.Request)
}
