package role

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// ============================================================================
// NewGetAssignableRolesUseCase
// ============================================================================

func TestNewGetAssignableRolesUseCase(t *testing.T) {
	t.Run("when called then returns use case", func(t *testing.T) {
		// Arrange
		svc := mocks.NewRoleService(t)

		// Act
		uc := NewGetAssignableRolesUseCase(svc)

		// Assert
		assert.NotNil(t, uc)
	})
}

// ============================================================================
// Execute
// ============================================================================

func TestGetAssignableRolesUseCase_Execute(t *testing.T) {
	t.Run("when service returns roles then returns roles", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expected := []*models.Role{
			{ID: 2, Name: "admin"},
			{ID: 3, Name: "encargado"},
		}

		svcMock := mocks.NewRoleService(t)
		svcMock.EXPECT().GetAllAssignable(ctx).Return(expected, nil)

		uc := NewGetAssignableRolesUseCase(svcMock)

		// Act
		result, err := uc.Execute(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expectedErr := errors.New("service error")

		svcMock := mocks.NewRoleService(t)
		svcMock.EXPECT().GetAllAssignable(ctx).Return(nil, expectedErr)

		uc := NewGetAssignableRolesUseCase(svcMock)

		// Act
		result, err := uc.Execute(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}
