package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// ShopService contains business logic, validations, and data access coordination.
// Use Cases orchestrate the flow using this service.
type ShopService interface {
	// GetByID retrieves a shop by ID.
	GetByID(ctx context.Context, shopID int) (*models.Shop, error)

	// GetBySlug retrieves a shop by slug.
	// Used for public store endpoints.
	GetBySlug(ctx context.Context, slug string) (*models.Shop, error)

	// GetActivePaymentMethod retrieves an active payment method from the shop.
	// Returns ValidationError if not found or inactive.
	GetActivePaymentMethod(shop *models.Shop, paymentMethodID int) (*models.PaymentMethod, error)

	// GetActiveDeliveryMethod retrieves an active delivery method from the shop.
	// Returns ValidationError if not found or inactive.
	GetActiveDeliveryMethod(shop *models.Shop, deliveryMethodID int) (*models.DeliveryMethod, error)

	// ValidateShippingCost validates the shipping cost matches the delivery method configuration.
	// Returns ValidationError if mismatch.
	ValidateShippingCost(shippingCost float64, method *models.DeliveryMethod, selectedZoneID int) error

	// Update updates an existing shop with optional new images.
	// Validates shop, uploads new images to storage, persists via repository.
	// Handles cleanup of removed images from storage.
	// newLogoBuffer and newCoverBuffer are optional (nil if no new image).
	Update(ctx context.Context, shopID int, shop *models.Shop, newLogoBuffer []byte, newCoverBuffer []byte) error
}
