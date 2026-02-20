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
// UpdateOrderStatusUseCase Tests
// =============================================================================

func TestUpdateOrderStatusUseCase_Execute(t *testing.T) {
	t.Run("when valid transition then updates status and returns order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		newStatus := models.OrderStatusConfirmed
		existingOrder := &models.Order{
			ID:     1,
			Status: models.OrderStatusPending,
			Store:  &models.Store{ID: 1},
		}
		expectedOrder := &models.Order{
			ID:     1,
			Status: models.OrderStatusConfirmed,
			Store:  &models.Store{ID: 1},
		}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil).Once()
		orderServiceMock.EXPECT().
			UpdateStatus(ctx, shopID, existingOrder, newStatus).
			Return(nil)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil).Once()

		useCase := NewUpdateOrderStatusUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, newStatus)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, models.OrderStatusConfirmed, result.Status)
	})

	t.Run("when order not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 999
		newStatus := models.OrderStatusConfirmed
		notFoundError := &errors.RecordNotFoundError{Message: errors.OrderNotFound}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, notFoundError)

		useCase := NewUpdateOrderStatusUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when update status fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		newStatus := models.OrderStatusConfirmed
		existingOrder := &models.Order{
			ID:     1,
			Status: models.OrderStatusPending,
			Store:  &models.Store{ID: 1},
		}
		updateError := &errors.ValidationError{Message: errors.InvalidOrderTransition}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)
		orderServiceMock.EXPECT().
			UpdateStatus(ctx, shopID, existingOrder, newStatus).
			Return(updateError)

		useCase := NewUpdateOrderStatusUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when re-fetch fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		newStatus := models.OrderStatusConfirmed
		existingOrder := &models.Order{
			ID:     1,
			Status: models.OrderStatusPending,
			Store:  &models.Store{ID: 1},
		}
		dbError := stdErrors.New("database error")

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil).Once()
		orderServiceMock.EXPECT().
			UpdateStatus(ctx, shopID, existingOrder, newStatus).
			Return(nil)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, dbError).Once()

		useCase := NewUpdateOrderStatusUseCase(orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
	})
}
