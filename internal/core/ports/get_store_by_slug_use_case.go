package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetStoreBySlugUseCase orchestrates retrieving a store by its slug.
type GetStoreBySlugUseCase interface {
	Execute(ctx context.Context, slug string) (*models.Store, error)
}
