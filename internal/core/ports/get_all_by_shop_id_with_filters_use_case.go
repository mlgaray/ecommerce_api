package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type GetAllByShopIDWithFiltersUseCase interface {
	// Execute retrieves products with filters and cursor-based pagination
	// ShopID is a context parameter (not a filter), passed separately
	// Filters are passed by pointer so normalized values (Limit, SortBy, SortOrder) propagate back
	// Returns: products, nextCursor (opaque string), hasMore flag, totalCount (nil on subsequent pages), error
	Execute(ctx context.Context, shopID int, filters *models.ProductFilters) ([]*models.Product, string, bool, *int, error)
}
