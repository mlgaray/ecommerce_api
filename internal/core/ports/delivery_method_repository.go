package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// DeliveryMethodRepository defines the interface for delivery method data access
type DeliveryMethodRepository interface {
	// GetAll returns all active delivery methods from the catalog
	GetAll(ctx context.Context) ([]*models.DeliveryMethod, error)

	// GetByID retrieves a delivery method by its ID
	GetByID(ctx context.Context, id int) (*models.DeliveryMethod, error)

	// GetByCode retrieves a delivery method by its code
	GetByCode(ctx context.Context, code models.DeliveryMethodCode) (*models.DeliveryMethod, error)
}
