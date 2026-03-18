package postgresql

import (
	"context"
	stdErrors "errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func init() {
	logs.Init()
}

func TestNewShopRepository(t *testing.T) {
	t.Run("when called then returns ShopSQLRepository", func(t *testing.T) {
		// Arrange
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mockDbConnection := mocks.NewDataBaseConnection(t)
		mockDbConnection.EXPECT().Connect().Return(db)

		paymentMethodRepo := mocks.NewPaymentMethodRepository(t)
		deliveryMethodRepo := mocks.NewDeliveryMethodRepository(t)

		// Act
		repo := NewShopRepository(mockDbConnection, paymentMethodRepo, deliveryMethodRepo)

		// Assert
		assert.NotNil(t, repo)
		assert.IsType(t, &ShopSQLRepository{}, repo)
	})
}

func TestShopRepository_SlugExists(t *testing.T) {
	t.Run("when slug exists then returns true", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		slug := "existing-store"

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(slug).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		repo := &ShopSQLRepository{db: db}

		// Act
		exists, err := repo.SlugExists(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when slug does not exist then returns false", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		slug := "new-store"

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(slug).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		repo := &ShopSQLRepository{db: db}

		// Act
		exists, err := repo.SlugExists(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database error occurs then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		slug := "some-store"

		mock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(slug).
			WillReturnError(fmt.Errorf("connection refused"))

		repo := &ShopSQLRepository{db: db}

		// Act
		exists, err := repo.SlugExists(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "database operation failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// =============================================================================
// handleInsertShopError Tests
// =============================================================================

func TestShopRepository_HandleInsertShopError(t *testing.T) {
	t.Run("when slug unique violation then returns DuplicateRecordError", func(t *testing.T) {
		// Arrange
		repo := &ShopSQLRepository{}
		pqErr := &pq.Error{
			Code:       PqErrCodeUniqueViolation,
			Constraint: ConstraintShopSlugUnique,
		}

		// Act
		err := repo.handleInsertShopError(pqErr, "taken-slug")

		// Assert
		assert.Error(t, err)
		var dupErr *errors.DuplicateRecordError
		assert.True(t, stdErrors.As(err, &dupErr))
		assert.Equal(t, errors.SlugAlreadyExists, dupErr.Message)
	})

	t.Run("when unique violation on different constraint then returns generic error", func(t *testing.T) {
		// Arrange
		repo := &ShopSQLRepository{}
		pqErr := &pq.Error{
			Code:       PqErrCodeUniqueViolation,
			Constraint: "shops_email_key",
		}

		// Act
		err := repo.handleInsertShopError(pqErr, "some-slug")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database operation failed")
		var dupErr *errors.DuplicateRecordError
		assert.False(t, stdErrors.As(err, &dupErr))
	})

	t.Run("when non-pq error then returns generic error", func(t *testing.T) {
		// Arrange
		repo := &ShopSQLRepository{}
		genericErr := fmt.Errorf("unexpected database error")

		// Act
		err := repo.handleInsertShopError(genericErr, "some-slug")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database operation failed")
	})
}
