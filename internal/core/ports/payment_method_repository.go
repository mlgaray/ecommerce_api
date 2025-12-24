package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// PaymentMethodRepository defines the interface for payment method data access
type PaymentMethodRepository interface {
	// GetAll returns all active payment methods from the catalog
	GetAll(ctx context.Context) ([]*models.PaymentMethod, error)

	// GetByID retrieves a payment method by its ID
	GetByID(ctx context.Context, id int) (*models.PaymentMethod, error)

	// GetByCode retrieves a payment method by its code
	GetByCode(ctx context.Context, code models.PaymentMethodCode) (*models.PaymentMethod, error)
}
