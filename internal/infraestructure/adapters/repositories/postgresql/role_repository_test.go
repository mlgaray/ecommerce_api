package postgresql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRoleSQLRepository_GetByName(t *testing.T) {
	t.Run("when role exists with direct DB connection then returns role successfully", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		roleName := "admin"
		expectedRole := &models.Role{
			ID:          1,
			Name:        "admin",
			Description: "Administrator role",
		}

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).
				AddRow(expectedRole.ID, expectedRole.Name, expectedRole.Description))

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedRole, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when role exists with transaction then returns role successfully", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.WithValue(context.Background(), TxContextKey, tx)
		roleName := "user"
		expectedRole := &models.Role{
			ID:          2,
			Name:        "user",
			Description: "Regular user role",
		}

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).
				AddRow(expectedRole.ID, expectedRole.Name, expectedRole.Description))

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedRole, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when role does not exist with direct DB connection then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		roleName := "nonexistent"

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnError(sql.ErrNoRows)

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when role does not exist with transaction then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.WithValue(context.Background(), TxContextKey, tx)
		roleName := "nonexistent"

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnError(sql.ErrNoRows)

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database connection fails with direct DB then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		ctx := context.Background()
		roleName := "admin"
		expectedError := sql.ErrConnDone

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnError(expectedError)

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("when database connection fails with transaction then returns error", func(t *testing.T) {
		// Arrange
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		assert.NoError(t, err)

		ctx := context.WithValue(context.Background(), TxContextKey, tx)
		roleName := "admin"
		expectedError := sql.ErrTxDone

		expectedQuery := `SELECT id, name, description FROM roles WHERE name = \$1`
		mock.ExpectQuery(expectedQuery).
			WithArgs(roleName).
			WillReturnError(expectedError)

		repo := &RoleSQLRepository{db: db}

		// Act
		role, err := repo.GetByName(ctx, roleName)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewRoleRepository(t *testing.T) {
	t.Run("when called then returns RoleRepository", func(t *testing.T) {
		// Arrange
		db, _, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		mockDbConnection := mocks.NewDataBaseConnection(t)
		mockDbConnection.EXPECT().Connect().Return(db)

		// Act
		repo := NewRoleRepository(mockDbConnection)

		// Assert
		assert.NotNil(t, repo)
		assert.IsType(t, &RoleSQLRepository{}, repo)
	})
}
