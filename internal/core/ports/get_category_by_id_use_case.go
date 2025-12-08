package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetCategoryByIDUseCase defines the interface for retrieving a category by ID.
type GetCategoryByIDUseCase interface {
	Execute(ctx context.Context, categoryID int) (*models.Category, error)
}
