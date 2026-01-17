package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetStoreCategoriesUseCase orchestrates category retrieval by store slug.
// Resolves slug to shop ID and delegates to category services for data retrieval.
// Note: This use case does NOT return totalCount as the store frontend doesn't use pagination UI.
type GetStoreCategoriesUseCase interface {
	// Execute retrieves categories for a store identified by slug.
	// Returns:
	//   - categories: List of categories (max of 'limit' items)
	//   - nextCursor: Opaque cursor for next page (empty if no more pages)
	//   - hasMore: true if there are more pages
	//   - error: Any error that occurred (store_not_found, validation errors, etc.)
	Execute(ctx context.Context, slug string, filters models.CategoryFilters) ([]*models.Category, string, bool, error)
}
