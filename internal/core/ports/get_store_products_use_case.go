package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetStoreProductsUseCase orchestrates products retrieval by store slug.
// Resolves slug to shop ID and delegates to product services for data retrieval.
// Automatically filters by is_active=true (public storefront only shows active products).
// Supports all standard product filters for infinite scroll pagination.
// Note: This use case does NOT return totalCount as the store frontend doesn't need it.
type GetStoreProductsUseCase interface {
	// Execute retrieves products for a store identified by slug.
	// Returns:
	//   - products: List of products (max of 'limit' items)
	//   - nextCursor: Opaque cursor for next page (empty if no more pages)
	//   - hasMore: true if there are more pages
	//   - error: Any error that occurred (store_not_found, validation errors, etc.)
	Execute(ctx context.Context, slug string, filters models.ProductFilters) ([]*models.Product, string, bool, error)
}
