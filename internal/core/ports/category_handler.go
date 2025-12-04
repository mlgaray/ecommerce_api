package ports

import "net/http"

// CategoryHandler defines the HTTP handler interface for category endpoints.
type CategoryHandler interface {
	// Create handles POST /categories requests.
	Create(w http.ResponseWriter, r *http.Request)

	// GetAllByShopIDWithFilters handles GET /shops/{shop_id}/categories requests.
	// Supports search, sorting, and cursor-based pagination.
	GetAllByShopIDWithFilters(w http.ResponseWriter, r *http.Request)
}
