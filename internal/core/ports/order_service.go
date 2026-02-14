package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// OrderService contains business logic for order operations.
// Persists orders to the database.
type OrderService interface {
	// Create persists a new order.
	// Expects the order to be already validated by StoreService.
	// Sets store data from the store parameter.
	// Sets status to "pending".
	// Validates order business rules before persisting.
	Create(ctx context.Context, order *models.Order, store *models.Store) (*models.Order, error)

	// GetAllByShopIDWithFilters retrieves orders with filters.
	// Assumes filters are already validated by the Use Case.
	// Returns orders with LIMIT+1 strategy for pagination.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) ([]*models.Order, error)

	// CountByShopIDWithFilters returns total count of orders matching filters.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) (int, error)
}
