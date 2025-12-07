package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Category repository log field constants
const (
	CategoryRepositoryField             = "category_repository"
	CategoryCreateFunctionField         = "create"
	CategoryUpdateFunctionField         = "update"
	CategoryGetByIDFunctionField        = "get_by_id"
	CategoryGetAllByShopIDFunctionField = "get_all_by_shop_id_with_filters"
	CategoryCountByShopIDFunctionField  = "count_by_shop_id_with_filters"
)

// Category repository log message constants
const (
	LogCategoryAlreadyExists = "Category with name already exists in shop"
	LogCategoryNotFound      = "Category not found"
	LogFailedCreateCategory  = "Failed to create category"
	LogFailedUpdateCategory  = "Failed to update category"
	LogFailedGetCategory     = "Failed to get category"
	LogFailedQueryCategories = "Failed to query categories"
	LogFailedCountCategories = "Failed to count categories"
	LogFailedScanCategory    = "Failed to scan category row"
	LogFailedUnmarshalImage  = "Failed to unmarshal category image"
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
		if pqErr.Code == PqErrCodeUniqueViolation && pqErr.Constraint == ConstraintCategoryNameShopIDUnique {
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

// Update updates an existing category with optional new image.
// Uses stored procedure to update category + replace image atomically.
// Validates that the category belongs to the specified shop.
// Returns the storage_ref of the deleted image (if any) for cleanup.
func (r *CategoryRepository) Update(ctx context.Context, categoryID int, category *models.Category, shopID int) (string, error) {
	const query = `SELECT update_category($1, $2, $3, $4, $5, $6)`

	// Extract image data if present (new image to replace existing)
	var imageURL, storageRef string
	if category.Image != nil && category.Image.ID == 0 {
		// New image (ID = 0 means it's a new upload, not existing)
		imageURL = category.Image.URL
		storageRef = category.Image.StorageRef
	}

	var deletedRef string
	err := r.db.QueryRowContext(ctx, query,
		categoryID,
		shopID,
		category.Name,
		category.Description,
		imageURL,
		storageRef,
	).Scan(&deletedRef)

	if err != nil {
		// Check if it's a PostgreSQL error
		if pqErr, ok := err.(*pq.Error); ok {
			// Check for "Category not found" (P0003)
			if pqErr.Code == PqErrCodeRaiseException {
				logs.WithFields(map[string]interface{}{
					"file":        CategoryRepositoryField,
					"function":    CategoryUpdateFunctionField,
					"category_id": categoryID,
				}).Warn(LogCategoryNotFound)
				return "", &errors.RecordNotFoundError{Message: errors.CategoryNotFound}
			}

			// Unique constraint violation (name + shop_id)
			if pqErr.Code == PqErrCodeUniqueViolation && pqErr.Constraint == ConstraintCategoryNameShopIDUnique {
				logs.WithFields(map[string]interface{}{
					"file":          CategoryRepositoryField,
					"function":      CategoryUpdateFunctionField,
					"constraint":    pqErr.Constraint,
					"category_name": category.Name,
					"category_id":   categoryID,
				}).Error(LogCategoryAlreadyExists)
				return "", &errors.DuplicateRecordError{Message: errors.CategoryAlreadyExistsInShop}
			}
		}

		logs.WithFields(map[string]interface{}{
			"file":        CategoryRepositoryField,
			"function":    CategoryUpdateFunctionField,
			"category_id": categoryID,
			"error":       err.Error(),
		}).Error(LogFailedUpdateCategory)
		return "", fmt.Errorf("database operation failed")
	}

	logs.WithFields(map[string]interface{}{
		"file":           CategoryRepositoryField,
		"function":       CategoryUpdateFunctionField,
		"category_id":    categoryID,
		"image_replaced": deletedRef != "",
	}).Info("Category updated successfully")

	return deletedRef, nil
}

// GetByID retrieves a category by its ID.
// Returns RecordNotFoundError if category doesn't exist.
func (r *CategoryRepository) GetByID(ctx context.Context, categoryID int) (*models.Category, error) {
	query := `
		SELECT
			c.id, c.name, COALESCE(c.description, ''), c.created_at,
			COALESCE(
				(SELECT jsonb_build_object('id', img.id, 'url', img.url, 'storage_ref', img.storage_ref)
				FROM images img
				WHERE img.category_id = c.id
				LIMIT 1),
				'null'::jsonb
			) AS image
		FROM categories c
		WHERE c.id = $1`

	category := &models.Category{}
	var imageJSON []byte

	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&imageJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logs.WithFields(map[string]interface{}{
				"file":        CategoryRepositoryField,
				"function":    CategoryGetByIDFunctionField,
				"category_id": categoryID,
			}).Warn(LogCategoryNotFound)
			return nil, &errors.RecordNotFoundError{Message: errors.CategoryNotFound}
		}

		logs.WithFields(map[string]interface{}{
			"file":        CategoryRepositoryField,
			"function":    CategoryGetByIDFunctionField,
			"category_id": categoryID,
			"error":       err.Error(),
		}).Error(LogFailedGetCategory)
		return nil, fmt.Errorf("database operation failed")
	}

	// Parse image JSON (can be null)
	if string(imageJSON) != JSONNull {
		var image models.Image
		if err := json.Unmarshal(imageJSON, &image); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":        CategoryRepositoryField,
				"function":    CategoryGetByIDFunctionField,
				"category_id": categoryID,
				"error":       err.Error(),
			}).Error(LogFailedUnmarshalImage)
			return nil, fmt.Errorf("database operation failed")
		}
		category.Image = &image
	}

	return category, nil
}

// GetAllByShopIDWithFilters retrieves categories with filters and pagination.
// Returns categories with LIMIT+1 strategy for cursor-based pagination.
//
//nolint:gocyclo // Complexity is inherent to dynamic filter/pagination logic
func (r *CategoryRepository) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) ([]*models.Category, error) {
	startTime := time.Now()

	baseQuery := `
		SELECT
			c.id, c.name, COALESCE(c.description, ''), c.created_at,
			COALESCE(
				(SELECT jsonb_build_object('id', img.id, 'url', img.url)
				FROM images img
				WHERE img.category_id = c.id
				LIMIT 1),
				'null'::jsonb
			) AS image
		FROM categories c
		WHERE c.shop_id = $1`

	// Build dynamic WHERE conditions
	conditions := []string{}
	args := []interface{}{shopID}
	argPos := 2

	// Search filter (by name)
	if filters.Search != nil && *filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("c.name ILIKE $%d", argPos))
		args = append(args, "%"+*filters.Search+"%")
		argPos++
	}

	// Keyset pagination
	if filters.LastID != nil {
		sortField := fmt.Sprintf("c.%s", filters.SortBy)

		if filters.SortOrder == SortOrderDesc {
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s < $%d OR (%s = $%d AND c.id < $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				conditions = append(conditions, fmt.Sprintf("c.id < $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		} else {
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s > $%d OR (%s = $%d AND c.id > $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				conditions = append(conditions, fmt.Sprintf("c.id > $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		}
	}

	// Append WHERE conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// ORDER BY
	baseQuery += fmt.Sprintf(" ORDER BY c.%s %s, c.id %s",
		filters.SortBy,
		strings.ToUpper(filters.SortOrder),
		strings.ToUpper(filters.SortOrder))

	// LIMIT+1 for pagination detection
	baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, filters.Limit+1)

	logs.WithFields(map[string]interface{}{
		"file":       CategoryRepositoryField,
		"function":   CategoryGetAllByShopIDFunctionField,
		"shop_id":    shopID,
		"sort_by":    filters.SortBy,
		"sort_order": filters.SortOrder,
		"has_search": filters.Search != nil,
		"limit":      filters.Limit,
		"last_id":    filters.LastID,
	}).Debug("Executing category query with filters")

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryRepositoryField,
			"function":    CategoryGetAllByShopIDFunctionField,
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		}).Error(LogFailedQueryCategories)
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	categories := make([]*models.Category, 0, filters.Limit)

	for rows.Next() {
		category := &models.Category{}
		var imageJSON []byte

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&imageJSON,
		)

		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     CategoryRepositoryField,
				"function": CategoryGetAllByShopIDFunctionField,
				"error":    err.Error(),
			}).Error(LogFailedScanCategory)
			return nil, fmt.Errorf("database operation failed")
		}

		// Parse image JSON (can be null)
		if string(imageJSON) != JSONNull {
			var image models.Image
			if err := json.Unmarshal(imageJSON, &image); err != nil {
				logs.WithFields(map[string]interface{}{
					"file":        CategoryRepositoryField,
					"function":    CategoryGetAllByShopIDFunctionField,
					"category_id": category.ID,
					"error":       err.Error(),
				}).Error(LogFailedUnmarshalImage)
				return nil, fmt.Errorf("database operation failed")
			}
			category.Image = &image
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryRepositoryField,
			"function": CategoryGetAllByShopIDFunctionField,
			"error":    err.Error(),
		}).Error("Error iterating category rows")
		return nil, fmt.Errorf("database operation failed")
	}

	logs.WithFields(map[string]interface{}{
		"file":         CategoryRepositoryField,
		"function":     CategoryGetAllByShopIDFunctionField,
		"duration_ms":  time.Since(startTime).Milliseconds(),
		"result_count": len(categories),
		"shop_id":      shopID,
	}).Debug("Categories query completed")

	return categories, nil
}

// CountByShopIDWithFilters returns total count of categories matching filters.
func (r *CategoryRepository) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CategoryFilters) (int, error) {
	startTime := time.Now()

	countQuery := "SELECT COUNT(*) FROM categories c WHERE c.shop_id = $1"
	args := []interface{}{shopID}
	argPos := 2

	var conditions []string

	// Search filter
	if filters.Search != nil && *filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("c.name ILIKE $%d", argPos))
		args = append(args, "%"+*filters.Search+"%")
	}

	if len(conditions) > 0 {
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryRepositoryField,
			"function":    CategoryCountByShopIDFunctionField,
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		}).Error(LogFailedCountCategories)
		return 0, fmt.Errorf("database operation failed")
	}

	logs.WithFields(map[string]interface{}{
		"file":        CategoryRepositoryField,
		"function":    CategoryCountByShopIDFunctionField,
		"duration_ms": time.Since(startTime).Milliseconds(),
		"total_count": totalCount,
		"shop_id":     shopID,
	}).Debug("COUNT query completed")

	return totalCount, nil
}
