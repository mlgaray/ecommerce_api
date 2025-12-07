package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// UpdateCategoryUseCase defines the interface for updating a category.
type UpdateCategoryUseCase interface {
	// Execute updates a category with optional new image.
	// If newImageBuffer is provided, the existing image is replaced.
	// Returns the deleted storage_ref for cleanup if image was replaced.
	Execute(ctx context.Context, categoryID int, category *models.Category, newImageBuffer []byte, shopID int) error
}
