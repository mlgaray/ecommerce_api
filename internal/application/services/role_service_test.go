package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// ============================================================================
// NewRoleService
// ============================================================================

func TestNewRoleService(t *testing.T) {
	t.Run("when called then returns RoleServiceImpl", func(t *testing.T) {
		// Arrange
		repo := mocks.NewRoleRepository(t)

		// Act
		svc := NewRoleService(repo)

		// Assert
		assert.NotNil(t, svc)
		assert.IsType(t, &RoleServiceImpl{}, svc)
	})
}

// ============================================================================
// GetAllAssignable
// ============================================================================

func TestRoleServiceImpl_GetAllAssignable(t *testing.T) {
	t.Run("when repository returns roles then returns roles", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expected := []*models.Role{
			{ID: 2, Name: "admin", Description: "Administrador de la tienda con permisos operations completos"},
			{ID: 3, Name: "encargado", Description: "Encargado de tienda con permisos de products, categorías y órdenes"},
		}

		repo := mocks.NewRoleRepository(t)
		repo.EXPECT().GetAllAssignable(ctx).Return(expected, nil)

		svc := NewRoleService(repo)

		// Act
		result, err := svc.GetAllAssignable(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when repository returns no roles then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		repo := mocks.NewRoleRepository(t)
		repo.EXPECT().GetAllAssignable(ctx).Return([]*models.Role{}, nil)

		svc := NewRoleService(repo)

		// Act
		result, err := svc.GetAllAssignable(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expectedErr := errors.New("database error")

		repo := mocks.NewRoleRepository(t)
		repo.EXPECT().GetAllAssignable(ctx).Return(nil, expectedErr)

		svc := NewRoleService(repo)

		// Act
		result, err := svc.GetAllAssignable(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}
