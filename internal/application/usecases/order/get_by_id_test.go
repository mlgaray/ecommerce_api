package order

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// GetOrderByIDUseCase Tests
// =============================================================================

func TestGetOrderByIDUseCase_Execute(t *testing.T) {
	t.Run("when order exists then returns order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		expectedOrder := &models.Order{
			ID:          1,
			OrderNumber: "ORD-001",
			Status:      models.OrderStatusPending,
			Store:       &models.Store{ID: 1, Name: "Test Store"},
		}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil)

		useCase := NewGetOrderByIDUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedOrder.ID, result.ID)
		assert.Equal(t, expectedOrder.OrderNumber, result.OrderNumber)
	})

	t.Run("when order not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 999
		notFoundError := &errors.RecordNotFoundError{Message: errors.OrderNotFound}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, notFoundError)

		useCase := NewGetOrderByIDUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		dbError := stdErrors.New("database error")

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, dbError)

		useCase := NewGetOrderByIDUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
	})
}
