package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// ProductService contains business logic, validations, and data access coordination.
// Use Cases orchestrate the flow using this service.
// Internal operations (validation, image preparation) are encapsulated within each method.
type ProductService interface {
	// Create creates a new product with images.
	// Validates product, prepares images, and persists via repository.
	Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error)

	// GetByID retrieves a product by ID.
	GetByID(ctx context.Context, productID int) (*models.Product, error)

	// Update updates an existing product with new images.
	// Validates product, uploads new images to storage, persists via repository.
	// Handles cleanup of removed images from storage.
	Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte, shopID int) error

	// GetAllByShopIDWithFilters retrieves products with filters.
	// Assumes filters are already validated by the Use Case.
	// Returns products with LIMIT+1 strategy for pagination.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) ([]*models.Product, error)

	// CountByShopIDWithFilters returns total count of products matching filters.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) (int, error)
}
