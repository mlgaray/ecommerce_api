package postgresql

import (
	"context"
	stdErrors "errors"
	"testing"

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
