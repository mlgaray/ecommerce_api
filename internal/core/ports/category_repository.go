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

	// Update updates an existing category with optional new image.
	// Uses stored procedure to update category + replace image atomically.
	// Validates that the category belongs to the specified shop.
	// Returns the storage_ref of the deleted image (if any) for cleanup.
	// Returns RecordNotFoundError if category doesn't exist or doesn't belong to shop.
	Update(ctx context.Context, categoryID int, category *models.Category, shopID int) (string, error)

	// GetByID retrieves a category by its ID.
	// Returns RecordNotFoundError if category doesn't exist.
	GetByID(ctx context.Context, categoryID int) (*models.Category, error)

	// GetAllByShopIDWithFilters retrieves categories with filters and pagination.
	// Returns categories with LIMIT+1 strategy for cursor-based pagination.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) ([]*models.Category, error)

	// CountByShopIDWithFilters returns total count of categories matching filters.
	// Used for first page to show "X of Y" in frontend.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) (int, error)
}
