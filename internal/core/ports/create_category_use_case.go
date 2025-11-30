package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CreateCategoryUseCase defines the use case for creating a new category.
// It acts as a coordinator, delegating business logic to the CategoryService.
type CreateCategoryUseCase interface {
	// Execute creates a new category.
	// Delegates to CategoryService for business logic and validation.
	Execute(ctx context.Context, category *models.Category, imageBuffer []byte, shopID int) (*models.Category, error)
}
