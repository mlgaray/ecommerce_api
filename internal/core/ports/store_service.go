package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// StoreService contains business logic for public store operations.
// Maps Shop data to Store model and handles store-specific calculations.
type StoreService interface {
	// GetBySlug retrieves a store by its slug.
	// Fetches shop data from repository and maps it to Store model.
	GetBySlug(ctx context.Context, slug string) (*models.Store, error)
}
