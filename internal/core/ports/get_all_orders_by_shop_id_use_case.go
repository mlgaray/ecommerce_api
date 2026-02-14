package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetAllOrdersByShopIDUseCase orchestrates order retrieval with filters and pagination.
// Coordinates OrderService and PaginationService.
type GetAllOrdersByShopIDUseCase interface {
	// Execute retrieves orders with filters and cursor-based pagination.
	// ShopID is a context parameter (not a filter), passed separately.
	//
	// Returns:
	//   - orders: List of orders (max of 'limit' items)
	//   - nextCursor: Opaque cursor for next page (empty if no more pages)
	//   - hasMore: true if there are more pages
	//   - totalCount: Total count of items (only on first page when LastID is nil, otherwise nil)
	//   - error: Any error that occurred
	Execute(ctx context.Context, shopID int, filters models.OrderFilters) (
		orders []*models.Order,
		nextCursor string,
		hasMore bool,
		totalCount *int,
		err error,
	)
}
