package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Category repository log field constants
const (
	CategoryRepositoryField     = "category_repository"
	CategoryCreateFunctionField = "create"
)

// Category repository log message constants
const (
	LogCategoryAlreadyExists = "Category with name already exists in shop"
	LogFailedCreateCategory  = "Failed to create category"
)

// CategoryRepository handles category data access operations.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository instance.
func NewCategoryRepository(dataBaseConnection DataBaseConnection) *CategoryRepository {
	return &CategoryRepository{
		db: dataBaseConnection.Connect(),
	}
}

// handlePostgreSQLError translates PostgreSQL errors to domain errors.
func (r *CategoryRepository) handlePostgreSQLError(err error, categoryName string, shopID int) error {
	if pqErr, ok := err.(*pq.Error); ok {
		// Unique constraint violation (name + shop_id)
		if pqErr.Code == "23505" && pqErr.Constraint == "categories_name_shop_id_unique" {
			logs.WithFields(map[string]interface{}{
				"file":          CategoryRepositoryField,
				"function":      CategoryCreateFunctionField,
				"constraint":    pqErr.Constraint,
				"category_name": categoryName,
				"shop_id":       shopID,
			}).Error(LogCategoryAlreadyExists)

			return &errors.DuplicateRecordError{
				Message: errors.CategoryAlreadyExistsInShop,
			}
		}
	}

	// Technical error - log details but return generic error
	logs.WithFields(map[string]interface{}{
		"file":          CategoryRepositoryField,
		"function":      CategoryCreateFunctionField,
		"category_name": categoryName,
		"shop_id":       shopID,
		"error":         err.Error(),
	}).Error(LogFailedCreateCategory)

	return fmt.Errorf("database operation failed")
}

// Create creates a new category with image in the database using stored procedure.
// Image data is read from category.Image if present.
// Returns DuplicateRecordError if a category with the same name already exists in the shop.
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category, shopID int) (*models.Category, error) {
	const query = `SELECT create_category($1, $2, $3, $4, $5)`

	// Extract image data if present
	var imageURL, storageRef string
	if category.Image != nil {
		imageURL = category.Image.URL
		storageRef = category.Image.StorageRef
	}

	var categoryID int
	err := r.db.QueryRowContext(ctx, query,
		category.Name,
		category.Description,
		shopID,
		imageURL,
		storageRef,
	).Scan(&categoryID)

	if err != nil {
		return nil, r.handlePostgreSQLError(err, category.Name, shopID)
	}

	category.ID = categoryID

	return category, nil
}
