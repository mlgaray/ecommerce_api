package ports

import "net/http"

// CategoryHandler defines the HTTP handler interface for category endpoints.
type CategoryHandler interface {
	// Create handles POST /categories requests.
	Create(w http.ResponseWriter, r *http.Request)

	// Update handles PUT /categories/{category_id} requests.
	// Supports optional image replacement.
	Update(w http.ResponseWriter, r *http.Request)

	// GetByID handles GET /categories/{category_id} requests.
	GetByID(w http.ResponseWriter, r *http.Request)

	// GetAllByShopIDWithFilters handles GET /shops/{shop_id}/categories requests.
	// Supports search, sorting, and cursor-based pagination.
	GetAllByShopIDWithFilters(w http.ResponseWriter, r *http.Request)

	// Delete handles DELETE /categories/{category_id} requests.
	// Returns 204 No Content on success.
	// Returns 404 if category doesn't exist or doesn't belong to shop.
	// Returns 409 if category has products assigned.
	Delete(w http.ResponseWriter, r *http.Request)
}
