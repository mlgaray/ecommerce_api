package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetStoreProductByIDUseCase orchestrates retrieving a product by ID within a store.
// Resolves slug to shop ID and validates product belongs to that shop.
// Returns product only if it exists, belongs to the store, and is active.
type GetStoreProductByIDUseCase interface {
	// Execute retrieves a product by ID for a store identified by slug.
	// Returns:
	//   - product: The product with variants and options
	//   - error: store_not_found, product_not_found, or other errors
	Execute(ctx context.Context, slug string, productID int) (*models.Product, error)
}
