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

	// GetByID retrieves a single order with full details (items, selected options).
	// Uses CTE-based query with JSON aggregation for items and selections.
	// Filters by both order_id AND store_id for security.
	// Returns RecordNotFoundError if not found.
	GetByID(ctx context.Context, shopID int, orderID int) (*models.Order, error)

	// UpdateStatus updates the order status using optimistic locking.
	// The currentStatus is used in the WHERE clause to detect concurrent modifications.
	// Returns ConcurrentModificationError if 0 rows affected (status changed between read and write).
	UpdateStatus(ctx context.Context, shopID int, orderID int, currentStatus models.OrderStatus, newStatus models.OrderStatus) error

	// Update updates order fields and replaces all items within a transaction.
	// Uses stored procedure for single round-trip performance.
	// Updates: customer, payment method, delivery method, delivery zone, totals, updated_at.
	// Deletes all existing items (CASCADE handles selections) and inserts new items.
	// Filters by order_id AND store_id for security.
	// Returns RecordNotFoundError if order not found.
	Update(ctx context.Context, shopID int, order *models.Order) error
}
