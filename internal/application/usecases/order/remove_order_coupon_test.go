package order

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// RemoveOrderCouponUseCase Tests
// =============================================================================

func newOrderWithCoupon() *models.Order {
	return &models.Order{
		ID:          1,
		OrderNumber: "ORD-001",
		Status:      models.OrderStatusPending,
		Store: &models.Store{
			ID:   1,
			Name: "Test Store",
			Slug: "test-store",
		},
		Customer: &models.Customer{
			Name:  "John Doe",
			Phone: "123456789",
		},
		Items: []*models.OrderItem{
			{
				Product:   &models.Product{ID: 1, Name: "Product 1"},
				Quantity:  2,
				UnitPrice: 100.00,
			},
		},
		Coupon: &models.Coupon{
			ID:             1,
			Code:           "SAVE10",
			DiscountAmount: 20.00,
		},
		Subtotal:     200.00,
		ShippingCost: 10.00,
		Discount:     20.00,
		Total:        190.00,
	}
}

func newOrderWithoutCoupon() *models.Order {
	return &models.Order{
		ID:          1,
		OrderNumber: "ORD-001",
		Status:      models.OrderStatusPending,
		Store: &models.Store{
			ID:   1,
			Name: "Test Store",
			Slug: "test-store",
		},
		Customer: &models.Customer{
			Name:  "John Doe",
			Phone: "123456789",
		},
		Items: []*models.OrderItem{
			{
				Product:   &models.Product{ID: 1, Name: "Product 1"},
				Quantity:  2,
				UnitPrice: 100.00,
			},
		},
		Subtotal:     200.00,
		ShippingCost: 10.00,
		Total:        210.00,
	}
}

// =============================================================================
// Success Cases
// =============================================================================

func TestRemoveOrderCouponUseCase_Execute(t *testing.T) {
	t.Run("when order has coupon then removes coupon and recalculates total", func(t *testing.T) {
		// Arrange
		orderServiceMock := mocks.NewOrderService(t)
		uc := NewRemoveOrderCouponUseCase(orderServiceMock)

		existingOrder := newOrderWithCoupon()
		expectedOrder := newOrderWithoutCoupon()

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 1).
			Return(existingOrder, nil).Once()

		orderServiceMock.EXPECT().
			Update(mock.Anything, 1, mock.MatchedBy(func(o *models.Order) bool {
				return o.Coupon == nil && o.Discount == 0 && o.Total == 210.00
			})).
			Return(nil).Once()

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 1).
			Return(expectedOrder, nil).Once()

		// Act
		result, err := uc.Execute(context.Background(), 1, 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Coupon)
		assert.Equal(t, 0.0, result.Discount)
		assert.Equal(t, 210.00, result.Total)
	})

	t.Run("when order has no coupon then returns order unchanged", func(t *testing.T) {
		// Arrange
		orderServiceMock := mocks.NewOrderService(t)
		uc := NewRemoveOrderCouponUseCase(orderServiceMock)

		existingOrder := newOrderWithoutCoupon()

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 1).
			Return(existingOrder, nil).Once()

		// Act (no Update or second GetByID expected)
		result, err := uc.Execute(context.Background(), 1, 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Coupon)
		assert.Equal(t, 210.00, result.Total)
	})
}

// =============================================================================
// Error Cases
// =============================================================================

func TestRemoveOrderCouponUseCase_Execute_OrderNotFound(t *testing.T) {
	t.Run("when order not found then returns error", func(t *testing.T) {
		// Arrange
		orderServiceMock := mocks.NewOrderService(t)
		uc := NewRemoveOrderCouponUseCase(orderServiceMock)

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 999).
			Return(nil, &errors.RecordNotFoundError{Message: errors.OrderNotFound}).Once()

		// Act
		result, err := uc.Execute(context.Background(), 999, 1)

		// Assert
		assert.Nil(t, result)
		assert.Error(t, err)
		var notFoundErr *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFoundErr))
	})
}

func TestRemoveOrderCouponUseCase_Execute_NotEditable(t *testing.T) {
	t.Run("when order is not editable then returns error", func(t *testing.T) {
		// Arrange
		orderServiceMock := mocks.NewOrderService(t)
		uc := NewRemoveOrderCouponUseCase(orderServiceMock)

		completedOrder := newOrderWithCoupon()
		completedOrder.Status = models.OrderStatusCompleted

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 1).
			Return(completedOrder, nil).Once()

		// Act
		result, err := uc.Execute(context.Background(), 1, 1)

		// Assert
		assert.Nil(t, result)
		assert.Error(t, err)
		var bizErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &bizErr))
		assert.Equal(t, errors.OrderCannotBeEdited, bizErr.Message)
	})
}

func TestRemoveOrderCouponUseCase_Execute_UpdateFails(t *testing.T) {
	t.Run("when update fails then returns error", func(t *testing.T) {
		// Arrange
		orderServiceMock := mocks.NewOrderService(t)
		uc := NewRemoveOrderCouponUseCase(orderServiceMock)

		existingOrder := newOrderWithCoupon()

		orderServiceMock.EXPECT().
			GetByID(mock.Anything, 1, 1).
			Return(existingOrder, nil).Once()

		orderServiceMock.EXPECT().
			Update(mock.Anything, 1, mock.Anything).
			Return(stdErrors.New("database error")).Once()

		// Act
		result, err := uc.Execute(context.Background(), 1, 1)

		// Assert
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
	})
}
