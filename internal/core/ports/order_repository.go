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
}
