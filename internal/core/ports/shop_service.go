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

	// Update updates an existing shop with optional new images.
	// Validates shop, uploads new images to storage, persists via repository.
	// Handles cleanup of removed images from storage.
	// newLogoBuffer and newCoverBuffer are optional (nil if no new image).
	Update(ctx context.Context, shopID int, shop *models.Shop, newLogoBuffer []byte, newCoverBuffer []byte) error
}
