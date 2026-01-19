package ports

import "net/http"

// StoreHandler handles HTTP requests for public store endpoints.
// These endpoints are accessible without authentication.
type StoreHandler interface {
	// GetBySlug handles GET /stores/{slug}
	GetBySlug(http.ResponseWriter, *http.Request)

	// GetCategories handles GET /stores/{slug}/categories
	GetCategories(http.ResponseWriter, *http.Request)

	// GetProducts handles GET /stores/{slug}/products
	GetProducts(http.ResponseWriter, *http.Request)

	// GetFeaturedProducts handles GET /stores/{slug}/products/featured
	GetFeaturedProducts(http.ResponseWriter, *http.Request)

	// GetProductByID handles GET /stores/{slug}/products/{productId}
	GetProductByID(http.ResponseWriter, *http.Request)
}
