package services

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
// Test Helpers
// =============================================================================

func newTestOrder() *models.Order {
	return &models.Order{
		Customer: &models.Customer{
			Name:  "John Doe",
			Phone: "123456789",
			Email: "john@example.com",
		},
		PaymentMethod: &models.PaymentMethod{
			ID:   1,
			Code: "cash",
			Name: "Cash",
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID:   1,
			Code: "delivery",
			Name: "Delivery",
		},
		Items: []*models.OrderItem{
			{
				Product:   &models.Product{ID: 1, Name: "Product 1"},
				Quantity:  2,
				UnitPrice: 100.00,
			},
		},
		ShippingCost: 10.00,
		Subtotal:     200.00,
		Total:        210.00,
	}
}

func newTestStore() *models.Store {
	return &models.Store{
		ID:   1,
		Name: "Test Store",
		Slug: "test-store",
	}
}

// =============================================================================
// OrderService.Create Tests
// =============================================================================

func TestOrderService_Create(t *testing.T) {
	t.Run("when order is valid then persists and returns order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		store := newTestStore()
		expectedOrder := newTestOrder()
		expectedOrder.ID = 1
		expectedOrder.Status = models.OrderStatusPending

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Create(ctx, order).
			Return(expectedOrder, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedOrder.ID, result.ID)
		assert.Equal(t, models.OrderStatusPending, order.Status)
		assert.Equal(t, store.ID, order.Store.ID)
		assert.Equal(t, store.Name, order.Store.Name)
		assert.Equal(t, store.Slug, order.Store.Slug)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		store := newTestStore()
		expectedError := stdErrors.New("database error")

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Create(ctx, order).
			Return(nil, expectedError)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when order validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := &models.Order{
			// Invalid order: missing required fields
			Customer: nil,
			Items:    []*models.OrderItem{},
		}
		store := newTestStore()

		orderRepoMock := mocks.NewOrderRepository(t)
		// Repository should not be called because validation fails first

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
	})

	t.Run("when subtotal does not match then returns subtotal mismatch error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		order.Subtotal = 999.99 // Wrong subtotal (should be 200)
		store := newTestStore()

		orderRepoMock := mocks.NewOrderRepository(t)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.SubtotalMismatch, validationErr.Message)
	})

	t.Run("when total does not match then returns total mismatch error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		order.Subtotal = 200.00 // Correct subtotal
		order.Total = 999.99    // Wrong total (should be 210)
		store := newTestStore()

		orderRepoMock := mocks.NewOrderRepository(t)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.TotalMismatch, validationErr.Message)
	})

	t.Run("when totals are correct then calculates item total prices", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		store := newTestStore()

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Create(ctx, order).
			Return(order, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Verify TotalPrice was calculated: 100.00 * 2 = 200.00
		assert.Equal(t, 200.00, order.Items[0].TotalPrice)
	})

	t.Run("when multiple items with correct totals then passes validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		order := newTestOrder()
		order.Items = append(order.Items, &models.OrderItem{
			Product:   &models.Product{ID: 2, Name: "Product 2"},
			Quantity:  3,
			UnitPrice: 50.00,
		})
		order.Subtotal = 350.00 // 200 + 150
		order.Total = 360.00    // 350 + 10 shipping
		store := newTestStore()

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Create(ctx, order).
			Return(order, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.Create(ctx, order, store)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200.00, order.Items[0].TotalPrice)
		assert.Equal(t, 150.00, order.Items[1].TotalPrice)
	})
}
