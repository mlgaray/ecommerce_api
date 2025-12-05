package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetAllCategoriesByShopIDWithFiltersUseCase orchestrates category retrieval with filters and pagination.
type GetAllCategoriesByShopIDWithFiltersUseCase interface {
	// Execute retrieves categories with filters and cursor-based pagination.
	// shopID is the context parameter (multi-tenancy scope).
	// Filters are passed by value (immutable). The use case calls Validated() internally.
	// Returns: categories, nextCursor (opaque string), hasMore flag, totalCount (nil on subsequent pages), error.
	Execute(ctx context.Context, shopID int, filters models.CategoryFilters) ([]*models.Category, string, bool, *int, error)
}
