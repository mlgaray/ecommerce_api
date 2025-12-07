package postgresql

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestNewCategoryRepository(t *testing.T) {
	t.Run("when called then returns CategoryRepository", func(t *testing.T) {
		// Arrange
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mockDbConnection := mocks.NewDataBaseConnection(t)
		mockDbConnection.EXPECT().Connect().Return(db)

		// Act
		repo := NewCategoryRepository(mockDbConnection)

		// Assert
		assert.NotNil(t, repo)
		assert.IsType(t, &CategoryRepository{}, repo)
	})
}

func TestCategoryRepository_Create(t *testing.T) {
	t.Run("when category is created successfully then returns category with ID", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		category := &models.Category{
			Name:        "Hamburguesas",
			Description: "Deliciosas hamburguesas",
			Image: &models.Image{
				URL:        "https://cloudinary.com/image.jpg",
				StorageRef: "shop_1/categories/abc123",
			},
		}

		mock.ExpectQuery(`SELECT create_category`).
			WithArgs(
				category.Name,
				category.Description,
				shopID,
				category.Image.URL,
				category.Image.StorageRef,
			).
			WillReturnRows(sqlmock.NewRows([]string{"create_category"}).AddRow(1))

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.Create(ctx, category, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, category.Name, result.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category without image then creates with empty image fields", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		category := &models.Category{
			Name:        "Bebidas",
			Description: "Refrescos y jugos",
			Image:       nil, // No image
		}

		mock.ExpectQuery(`SELECT create_category`).
			WithArgs(
				category.Name,
				category.Description,
				shopID,
				"", // Empty URL
				"", // Empty StorageRef
			).
			WillReturnRows(sqlmock.NewRows([]string{"create_category"}).AddRow(2))

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.Create(ctx, category, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category name already exists in shop then returns DuplicateRecordError", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		category := &models.Category{
			Name:        "Hamburguesas",
			Description: "Otra descripción",
		}

		pqErr := &pq.Error{
			Code:       "23505",
			Constraint: "categories_name_shop_id_unique",
		}

		mock.ExpectQuery(`SELECT create_category`).
			WithArgs(
				category.Name,
				category.Description,
				shopID,
				"",
				"",
			).
			WillReturnError(pqErr)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.Create(ctx, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var duplicateErr *errors.DuplicateRecordError
		assert.True(t, stdErrors.As(err, &duplicateErr))
		assert.Equal(t, errors.CategoryAlreadyExistsInShop, duplicateErr.Message)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database connection fails then returns generic error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		category := &models.Category{
			Name:        "Postres",
			Description: "Dulces y postres",
		}

		expectedError := stdErrors.New("connection refused")
		mock.ExpectQuery(`SELECT create_category`).
			WithArgs(
				category.Name,
				category.Description,
				shopID,
				"",
				"",
			).
			WillReturnError(expectedError)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.Create(ctx, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when other PostgreSQL error then returns generic error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		shopID := 1
		category := &models.Category{
			Name:        "Test Category",
			Description: "Test Description",
		}

		// Foreign key violation (shop doesn't exist)
		pqErr := &pq.Error{
			Code:       "23503",
			Constraint: "categories_shop_id_fkey",
		}

		mock.ExpectQuery(`SELECT create_category`).
			WithArgs(
				category.Name,
				category.Description,
				shopID,
				"",
				"",
			).
			WillReturnError(pqErr)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.Create(ctx, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoryRepository_GetByID(t *testing.T) {
	t.Run("when category exists then returns category with image", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		imageJSON := `{"id": 5, "url": "https://cloudinary.com/image.jpg", "storage_ref": "shop_1/categories/abc123"}`

		rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "image"}).
			AddRow(categoryID, "Electronics", "Electronic products", createdAt, imageJSON)

		mock.ExpectQuery(`SELECT(.*)FROM categories`).
			WithArgs(categoryID).
			WillReturnRows(rows)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.GetByID(ctx, categoryID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, categoryID, result.ID)
		assert.Equal(t, "Electronics", result.Name)
		assert.NotNil(t, result.Image)
		assert.Equal(t, 5, result.Image.ID)
		assert.Equal(t, "https://cloudinary.com/image.jpg", result.Image.URL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category exists without image then returns category with nil image", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "image"}).
			AddRow(categoryID, "Electronics", "Electronic products", createdAt, "null")

		mock.ExpectQuery(`SELECT(.*)FROM categories`).
			WithArgs(categoryID).
			WillReturnRows(rows)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.GetByID(ctx, categoryID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, categoryID, result.ID)
		assert.Nil(t, result.Image)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 999

		mock.ExpectQuery(`SELECT(.*)FROM categories`).
			WithArgs(categoryID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "image"}))

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.GetByID(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFoundErr *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFoundErr))
		assert.Equal(t, errors.CategoryNotFound, notFoundErr.Message)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database connection fails then returns generic error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1

		expectedError := stdErrors.New("connection refused")
		mock.ExpectQuery(`SELECT(.*)FROM categories`).
			WithArgs(categoryID).
			WillReturnError(expectedError)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.GetByID(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when image JSON is invalid then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		invalidImageJSON := `{invalid json}`

		rows := sqlmock.NewRows([]string{"id", "name", "description", "created_at", "image"}).
			AddRow(categoryID, "Electronics", "Electronic products", createdAt, invalidImageJSON)

		mock.ExpectQuery(`SELECT(.*)FROM categories`).
			WithArgs(categoryID).
			WillReturnRows(rows)

		repo := &CategoryRepository{db: db}

		// Act
		result, err := repo.GetByID(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCategoryRepository_Update(t *testing.T) {
	t.Run("when category is updated successfully with new image then returns old storage_ref", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		oldStorageRef := "shop_1/categories/old_ref"
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
			Image: &models.Image{
				ID:         0, // New image (ID = 0)
				URL:        "https://cloudinary.com/new_image.jpg",
				StorageRef: "shop_1/categories/new_ref",
			},
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				category.Image.URL,
				category.Image.StorageRef,
			).
			WillReturnRows(sqlmock.NewRows([]string{"update_category"}).AddRow(oldStorageRef))

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, oldStorageRef, deletedRef)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category is updated without new image then passes empty image fields", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
			Image: &models.Image{
				ID:  5, // Existing image (ID > 0)
				URL: "https://cloudinary.com/existing.jpg",
			},
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"", // Empty URL (existing image, not new)
				"", // Empty StorageRef
			).
			WillReturnRows(sqlmock.NewRows([]string{"update_category"}).AddRow(""))

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, deletedRef)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 999
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}

		pqErr := &pq.Error{
			Code:    "P0003", // Category not found
			Message: "Category not found or does not belong to shop",
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"",
				"",
			).
			WillReturnError(pqErr)

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, deletedRef)
		var notFoundErr *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFoundErr))
		assert.Equal(t, errors.CategoryNotFound, notFoundErr.Message)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category name already exists in shop then returns DuplicateRecordError", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Existing Name",
			Description: "Updated Description",
		}

		pqErr := &pq.Error{
			Code:       "23505",
			Constraint: "categories_name_shop_id_unique",
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"",
				"",
			).
			WillReturnError(pqErr)

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, deletedRef)
		var duplicateErr *errors.DuplicateRecordError
		assert.True(t, stdErrors.As(err, &duplicateErr))
		assert.Equal(t, errors.CategoryAlreadyExistsInShop, duplicateErr.Message)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database connection fails then returns generic error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}

		expectedError := stdErrors.New("connection refused")
		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"",
				"",
			).
			WillReturnError(expectedError)

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, deletedRef)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when other PostgreSQL error then returns generic error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}

		pqErr := &pq.Error{
			Code:    "P0001", // Generic raise exception
			Message: "Some other error",
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"",
				"",
			).
			WillReturnError(pqErr)

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, deletedRef)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when category has nil image then passes empty image fields", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
			Image:       nil, // No image
		}

		mock.ExpectQuery(`SELECT update_category`).
			WithArgs(
				categoryID,
				shopID,
				category.Name,
				category.Description,
				"",
				"",
			).
			WillReturnRows(sqlmock.NewRows([]string{"update_category"}).AddRow(""))

		repo := &CategoryRepository{db: db}

		// Act
		deletedRef, err := repo.Update(ctx, categoryID, category, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, deletedRef)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
