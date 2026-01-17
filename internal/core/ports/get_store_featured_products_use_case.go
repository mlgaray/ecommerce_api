package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetStoreFeaturedProductsUseCase orchestrates featured products retrieval by store slug.
// Resolves slug to shop ID and delegates to product services for data retrieval.
// Automatically filters by is_highlighted=true and is_active=true.
// Note: This use case does NOT return totalCount as the store frontend doesn't need it.
type GetStoreFeaturedProductsUseCase interface {
	// Execute retrieves featured products for a store identified by slug.
	// Returns:
	//   - products: List of featured products (max of 'limit' items)
	//   - nextCursor: Opaque cursor for next page (empty if no more pages)
	//   - hasMore: true if there are more pages
	//   - error: Any error that occurred (store_not_found, validation errors, etc.)
	Execute(ctx context.Context, slug string, filters models.ProductFilters) ([]*models.Product, string, bool, error)
}
