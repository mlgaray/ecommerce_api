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

// =============================================================================
// OrderService.GetAllByShopIDWithFilters Tests
// =============================================================================

func TestOrderService_GetAllByShopIDWithFilters(t *testing.T) {
	t.Run("when repository returns orders then returns orders", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.OrderFilters{Limit: 20, SortBy: "created_at", SortOrder: "desc"}
		expectedOrders := []*models.Order{
			{ID: 1, OrderNumber: "ORD-001", Total: 210.00},
			{ID: 2, OrderNumber: "ORD-002", Total: 150.00},
		}

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(expectedOrders, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, result)
		assert.Len(t, result, 2)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.OrderFilters{Limit: 20, SortBy: "created_at", SortOrder: "desc"}
		expectedError := stdErrors.New("database error")

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(nil, expectedError)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// OrderService.CountByShopIDWithFilters Tests
// =============================================================================

func TestOrderService_CountByShopIDWithFilters(t *testing.T) {
	t.Run("when repository returns count then returns count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.OrderFilters{Limit: 20, SortBy: "created_at", SortOrder: "desc"}
		expectedCount := 42

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, filters).
			Return(expectedCount, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, result)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.OrderFilters{Limit: 20, SortBy: "created_at", SortOrder: "desc"}
		expectedError := stdErrors.New("database error")

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, filters).
			Return(0, expectedError)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// OrderService.GetByID Tests
// =============================================================================

func TestOrderService_GetByID(t *testing.T) {
	t.Run("when order exists then returns order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		expectedOrder := &models.Order{ID: 1, OrderNumber: "ORD-001"}

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.GetByID(ctx, shopID, orderID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, result)
	})

	t.Run("when order not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 999
		notFoundError := &errors.RecordNotFoundError{Message: errors.OrderNotFound}

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, notFoundError)

		service := NewOrderService(orderRepoMock)

		// Act
		result, err := service.GetByID(ctx, shopID, orderID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// OrderService.UpdateStatus Tests
// =============================================================================

func TestOrderService_UpdateStatus(t *testing.T) {
	t.Run("when valid transition then updates status", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		order := &models.Order{
			ID:     1,
			Status: models.OrderStatusPending,
		}
		newStatus := models.OrderStatusConfirmed

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			UpdateStatus(ctx, shopID, order.ID, models.OrderStatusPending, newStatus).
			Return(nil)

		service := NewOrderService(orderRepoMock)

		// Act
		err := service.UpdateStatus(ctx, shopID, order, newStatus)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, models.OrderStatusConfirmed, order.Status)
	})

	t.Run("when invalid transition then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		order := &models.Order{
			ID:     1,
			Status: models.OrderStatusCompleted,
		}
		newStatus := models.OrderStatusPending // Invalid: completed -> pending

		orderRepoMock := mocks.NewOrderRepository(t)
		service := NewOrderService(orderRepoMock)

		// Act
		err := service.UpdateStatus(ctx, shopID, order, newStatus)

		// Assert
		assert.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.InvalidOrderTransition, valErr.Message)
	})

	t.Run("when concurrent modification then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		order := &models.Order{
			ID:     1,
			Status: models.OrderStatusPending,
		}
		newStatus := models.OrderStatusConfirmed
		concurrentError := &errors.ConcurrentModificationError{Message: errors.OrderConcurrentModification}

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			UpdateStatus(ctx, shopID, order.ID, models.OrderStatusPending, newStatus).
			Return(concurrentError)

		service := NewOrderService(orderRepoMock)

		// Act
		err := service.UpdateStatus(ctx, shopID, order, newStatus)

		// Assert
		assert.Error(t, err)
		var concErr *errors.ConcurrentModificationError
		assert.True(t, stdErrors.As(err, &concErr))
	})
}

// =============================================================================
// OrderService.Update Tests
// =============================================================================

func TestOrderService_Update(t *testing.T) {
	t.Run("when update succeeds then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		order := newTestOrder()
		order.ID = 1

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Update(ctx, shopID, order).
			Return(nil)

		service := NewOrderService(orderRepoMock)

		// Act
		err := service.Update(ctx, shopID, order)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		order := newTestOrder()
		order.ID = 1
		dbError := stdErrors.New("database error")

		orderRepoMock := mocks.NewOrderRepository(t)
		orderRepoMock.EXPECT().
			Update(ctx, shopID, order).
			Return(dbError)

		service := NewOrderService(orderRepoMock)

		// Act
		err := service.Update(ctx, shopID, order)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
	})
}

// =============================================================================
// OrderService.CalculateAndValidateTotals Tests (standalone)
// =============================================================================

func TestOrderService_CalculateAndValidateTotals(t *testing.T) {
	t.Run("when totals are correct then returns nil", func(t *testing.T) {
		// Arrange
		order := newTestOrder()
		orderRepoMock := mocks.NewOrderRepository(t)
		service := NewOrderService(orderRepoMock)

		// Act
		err := service.CalculateAndValidateTotals(order)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200.00, order.Items[0].TotalPrice)
	})

	t.Run("when subtotal mismatches then returns error", func(t *testing.T) {
		// Arrange
		order := newTestOrder()
		order.Subtotal = 999.99
		orderRepoMock := mocks.NewOrderRepository(t)
		service := NewOrderService(orderRepoMock)

		// Act
		err := service.CalculateAndValidateTotals(order)

		// Assert
		assert.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.SubtotalMismatch, valErr.Message)
	})

	t.Run("when total mismatches then returns error", func(t *testing.T) {
		// Arrange
		order := newTestOrder()
		order.Total = 999.99
		orderRepoMock := mocks.NewOrderRepository(t)
		service := NewOrderService(orderRepoMock)

		// Act
		err := service.CalculateAndValidateTotals(order)

		// Assert
		assert.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.TotalMismatch, valErr.Message)
	})
}
