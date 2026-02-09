package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// StoreService contains business logic for public store operations.
// Maps Shop data to Store model and handles store-specific calculations.
// Used by CreateOrderUseCase for order validation and creation.
type StoreService interface {
	// GetBySlug retrieves a store by its slug.
	// Fetches shop data from repository and maps it to Store model.
	GetBySlug(ctx context.Context, slug string) (*models.Store, error)

	// ValidateOrderItems validates that order items match current database data.
	// Compares product name, price, promotional price, variants and options.
	// Validates stock availability if product is stockeable.
	// Returns ValidationError or BusinessRuleError if validation fails.
	ValidateOrderItems(ctx context.Context, items []*models.OrderItem, storeID int) error

	// ValidateDeliveryMethod validates that the delivery method is active and shipping cost matches.
	// Returns ValidationError if method is not found, inactive, or price mismatch.
	ValidateDeliveryMethod(store *models.Store, deliveryMethod *models.DeliveryMethod, shippingCost float64) error
}
