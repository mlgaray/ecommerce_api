package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// OrderRepository handles order persistence operations.
type OrderRepository interface {
	// Create creates a new order with all its items and selected options.
	// Uses stored procedure to handle transactional insert.
	// Returns the created order with ID and order_number.
	Create(ctx context.Context, order *models.Order) (*models.Order, error)

	// GetAllByShopIDWithFilters returns orders for a shop with filters applied.
	// Lightweight query - NO items included (use GetByID for full details).
	// Returns limit+1 items for pagination (LIMIT+1 strategy).
	// Populates Order.ItemsCount with the count of items per order.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) ([]*models.Order, error)

	// CountByShopIDWithFilters returns total count of orders matching filters.
	// Used only on first page to show total count to user.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) (int, error)
}
