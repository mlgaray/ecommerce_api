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
// UpdateOrderUseCase Tests
// =============================================================================

func newEditableOrder() *models.Order {
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
			Email: "john@example.com",
		},
		PaymentMethod: &models.PaymentMethod{
			ID: 1, Name: "Cash", Code: "cash",
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID: 1, Name: "Delivery", Code: "delivery",
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

func newUpdatedData() *models.Order {
	return &models.Order{
		Customer: &models.Customer{
			Name:  "Jane Doe",
			Phone: "987654321",
			Email: "jane@example.com",
		},
		PaymentMethod: &models.PaymentMethod{
			ID: 2, Name: "Card", Code: "card",
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID: 1, Name: "Delivery", Code: "delivery",
		},
		Items: []*models.OrderItem{
			{
				Product:   &models.Product{ID: 1, Name: "Product 1"},
				Quantity:  3,
				UnitPrice: 100.00,
			},
		},
		ShippingCost: 10.00,
		Subtotal:     300.00,
		Total:        310.00,
	}
}

func TestUpdateOrderUseCase_Execute(t *testing.T) {
	t.Run("when all validations pass then updates and returns order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		expectedOrder := newEditableOrder()
		expectedOrder.Customer = updatedData.Customer

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil).Once()
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(existingOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			Update(ctx, shopID, existingOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil).Once()

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, existingOrder.Store.Slug).
			Return(newTestStore(), nil)
		storeServiceMock.EXPECT().
			ValidateDeliveryMethod(newTestStore(), updatedData.DeliveryMethod, updatedData.ShippingCost).
			Return(nil)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when order not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 999
		updatedData := newUpdatedData()
		notFoundError := &errors.RecordNotFoundError{Message: errors.OrderNotFound}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(nil, notFoundError)

		storeServiceMock := mocks.NewStoreService(t)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when order is not editable then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		completedOrder := newEditableOrder()
		completedOrder.Status = models.OrderStatusCompleted
		updatedData := newUpdatedData()

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(completedOrder, nil)

		storeServiceMock := mocks.NewStoreService(t)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &bizErr))
		assert.Equal(t, errors.OrderCannotBeEdited, bizErr.Message)
	})

	t.Run("when cancelled order then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		cancelledOrder := newEditableOrder()
		cancelledOrder.Status = models.OrderStatusCancelled
		updatedData := newUpdatedData()

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(cancelledOrder, nil)

		storeServiceMock := mocks.NewStoreService(t)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &bizErr))
		assert.Equal(t, errors.OrderCannotBeEdited, bizErr.Message)
	})

	t.Run("when order items validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		validationError := &errors.ValidationError{Message: errors.ProductDataMismatch}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(validationError)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when store not found for delivery validation then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		storeError := &errors.RecordNotFoundError{Message: errors.StoreNotFound}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, existingOrder.Store.Slug).
			Return(nil, storeError)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when delivery method validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		store := newTestStore()
		validationError := &errors.ValidationError{Message: errors.ShippingCostMismatch}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, existingOrder.Store.Slug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateDeliveryMethod(store, updatedData.DeliveryMethod, updatedData.ShippingCost).
			Return(validationError)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when totals validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		updatedData.DeliveryMethod = nil // Skip delivery validation
		validationError := &errors.ValidationError{Message: errors.SubtotalMismatch}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(existingOrder).
			Return(validationError)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.SubtotalMismatch, valErr.Message)
	})

	t.Run("when domain validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		updatedData.DeliveryMethod = nil
		updatedData.Customer = nil // Will cause domain validation failure
		updatedData.Items = []*models.OrderItem{}

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(existingOrder).
			Return(nil)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when persist fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		updatedData.DeliveryMethod = nil
		dbError := stdErrors.New("database error")

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil)
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(existingOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			Update(ctx, shopID, existingOrder).
			Return(dbError)

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
	})

	t.Run("when no delivery method then skips delivery validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		existingOrder := newEditableOrder()
		updatedData := newUpdatedData()
		updatedData.DeliveryMethod = nil
		expectedOrder := newEditableOrder()

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(existingOrder, nil).Once()
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(existingOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			Update(ctx, shopID, existingOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil).Once()

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, existingOrder.Store.ID).
			Return(nil)
		// GetBySlug and ValidateDeliveryMethod should NOT be called

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when confirmed order then allows editing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		orderID := 1
		confirmedOrder := newEditableOrder()
		confirmedOrder.Status = models.OrderStatusConfirmed
		updatedData := newUpdatedData()
		updatedData.DeliveryMethod = nil
		expectedOrder := newEditableOrder()

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(confirmedOrder, nil).Once()
		orderServiceMock.EXPECT().
			CalculateAndValidateTotals(confirmedOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			Update(ctx, shopID, confirmedOrder).
			Return(nil)
		orderServiceMock.EXPECT().
			GetByID(ctx, shopID, orderID).
			Return(expectedOrder, nil).Once()

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, updatedData.Items, confirmedOrder.Store.ID).
			Return(nil)

		useCase := NewUpdateOrderUseCase(orderServiceMock, storeServiceMock)

		// Act
		result, err := useCase.Execute(ctx, shopID, orderID, updatedData)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
