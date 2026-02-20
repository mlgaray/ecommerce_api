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

	// GetByID retrieves a single order with full details (items, selections).
	// Returns RecordNotFoundError if order doesn't exist or doesn't belong to shop.
	GetByID(ctx context.Context, shopID int, orderID int) (*models.Order, error)

	// UpdateStatus validates the status transition and persists the change.
	// Uses optimistic locking via the order's current status.
	// Returns ValidationError if transition is invalid.
	// Returns ConcurrentModificationError if the order was modified concurrently.
	UpdateStatus(ctx context.Context, shopID int, order *models.Order, newStatus models.OrderStatus) error

	// Update persists order changes. Expects the order to be already validated
	// by the use case (editability, items, delivery, totals, domain rules).
	// Does NOT change the order status.
	Update(ctx context.Context, shopID int, order *models.Order) error

	// CalculateAndValidateTotals calculates item total prices, subtotal, and total.
	// Validates that the frontend-sent values match the backend calculations.
	// Backend is the source of truth for all monetary calculations.
	CalculateAndValidateTotals(order *models.Order) error
}
