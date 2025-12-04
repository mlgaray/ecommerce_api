package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CategoryRepository defines the interface for category data access operations.
type CategoryRepository interface {
	// Create creates a new category with image in the database.
	// Uses stored procedure to create category + image atomically.
	// Image data (URL, StorageRef) is read from category.Image if present.
	// Returns DuplicateRecordError if a category with the same name already exists in the shop.
	Create(ctx context.Context, category *models.Category, shopID int) (*models.Category, error)

	// GetAllByShopIDWithFilters retrieves categories with filters and pagination.
	// Returns categories with LIMIT+1 strategy for cursor-based pagination.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) ([]*models.Category, error)

	// CountByShopIDWithFilters returns total count of categories matching filters.
	// Used for first page to show "X of Y" in frontend.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) (int, error)
}
