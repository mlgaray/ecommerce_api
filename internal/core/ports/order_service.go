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
}
