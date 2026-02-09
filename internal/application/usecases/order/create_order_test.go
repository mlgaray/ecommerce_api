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
			ID: 1,
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID: 1,
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
	fixedPrice := 10.00
	return &models.Store{
		ID:   1,
		Name: "Test Store",
		Slug: "test-store",
		PaymentMethods: []*models.PaymentMethod{
			{ID: 1, Name: "Cash", Code: "cash", IsActive: true},
		},
		DeliveryMethods: []*models.DeliveryMethod{
			{
				ID:       1,
				Name:     "Delivery",
				Code:     "delivery",
				IsActive: true,
				DeliveryConfig: &models.DeliveryConfig{
					FixedPrice: &fixedPrice,
				},
			},
		},
	}
}

// =============================================================================
// CreateOrderUseCase Tests
// =============================================================================

func TestCreateOrderUseCase_Execute(t *testing.T) {
	t.Run("when all validations pass then creates order", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		order := newTestOrder()
		store := newTestStore()
		expectedOrder := newTestOrder()
		expectedOrder.ID = 1

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, order.Items, store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			ValidateDeliveryMethod(store, order.DeliveryMethod, order.ShippingCost).
			Return(nil)

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			Create(ctx, order, store).
			Return(expectedOrder, nil)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedOrder.ID, result.ID)
	})

	t.Run("when store not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "non-existent-store"
		order := newTestOrder()
		notFoundError := &errors.RecordNotFoundError{Message: errors.StoreNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(nil, notFoundError)

		orderServiceMock := mocks.NewOrderService(t)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when order items validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		order := newTestOrder()
		store := newTestStore()
		validationError := &errors.ValidationError{Message: errors.ProductDataMismatch}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, order.Items, store.ID).
			Return(validationError)

		orderServiceMock := mocks.NewOrderService(t)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when delivery method validation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		order := newTestOrder()
		store := newTestStore()
		validationError := &errors.ValidationError{Message: errors.ShippingCostMismatch}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, order.Items, store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			ValidateDeliveryMethod(store, order.DeliveryMethod, order.ShippingCost).
			Return(validationError)

		orderServiceMock := mocks.NewOrderService(t)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})

	t.Run("when order service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		order := newTestOrder()
		store := newTestStore()
		expectedError := stdErrors.New("database error")

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, order.Items, store.ID).
			Return(nil)
		storeServiceMock.EXPECT().
			ValidateDeliveryMethod(store, order.DeliveryMethod, order.ShippingCost).
			Return(nil)

		orderServiceMock := mocks.NewOrderService(t)
		orderServiceMock.EXPECT().
			Create(ctx, order, store).
			Return(nil, expectedError)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when insufficient stock then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		order := newTestOrder()
		store := newTestStore()
		businessError := &errors.BusinessRuleError{Message: errors.InsufficientStock}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)
		storeServiceMock.EXPECT().
			ValidateOrderItems(ctx, order.Items, store.ID).
			Return(businessError)

		orderServiceMock := mocks.NewOrderService(t)

		useCase := NewCreateOrderUseCase(storeServiceMock, orderServiceMock)

		// Act
		result, err := useCase.Execute(ctx, order, storeSlug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &bizErr))
		assert.Equal(t, errors.InsufficientStock, bizErr.Message)
	})
}
