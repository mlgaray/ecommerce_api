package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CategoryService contains business logic for category operations.
// Handles validation, image upload orchestration, and persistence.
type CategoryService interface {
	// Create creates a new category with image.
	// Validates category, uploads image to storage, and persists via repository.
	// Handles rollback if persistence fails after image upload.
	Create(ctx context.Context, category *models.Category, imageBuffer []byte, shopID int) (*models.Category, error)
}
