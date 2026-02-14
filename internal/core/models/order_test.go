package models

import (
	stdErrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidOrder creates a valid order for testing
func newValidOrder() *Order {
	return &Order{
		Customer: &Customer{
			Name:  "John Doe",
			Phone: "123456789",
			Email: "john@example.com",
		},
		Items: []*OrderItem{
			{
				Product:  &Product{ID: 1, Name: "Test Product"},
				Quantity: 2,
			},
		},
	}
}

// =============================================================================
// Validate Tests - Customer
// =============================================================================

func TestOrder_Validate_Customer(t *testing.T) {
	t.Run("when customer is nil then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Customer = nil

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderCustomerNameRequired, validationErr.Message)
	})

	t.Run("when customer name is empty then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Customer.Name = ""

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderCustomerNameRequired, validationErr.Message)
	})

	t.Run("when customer name is valid then validation passes", func(t *testing.T) {
		order := newValidOrder()
		order.Customer.Name = "John Doe"

		err := order.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Items
// =============================================================================

func TestOrder_Validate_Items(t *testing.T) {
	t.Run("when items is empty then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Items = []*OrderItem{}

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderMustHaveItems, validationErr.Message)
	})

	t.Run("when items is nil then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Items = nil

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderMustHaveItems, validationErr.Message)
	})

	t.Run("when items have invalid item then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Items = []*OrderItem{
			{
				Product:  nil, // Invalid: no product
				Quantity: 2,
			},
		}

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})

	t.Run("when items are valid then validation passes", func(t *testing.T) {
		order := newValidOrder()

		err := order.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Multiple Items
// =============================================================================

func TestOrder_Validate_MultipleItems(t *testing.T) {
	t.Run("when all items are valid then validation passes", func(t *testing.T) {
		order := newValidOrder()
		order.Items = []*OrderItem{
			{Product: &Product{ID: 1}, Quantity: 1},
			{Product: &Product{ID: 2}, Quantity: 2},
			{Product: &Product{ID: 3}, Quantity: 3},
		}

		err := order.Validate()

		assert.NoError(t, err)
	})

	t.Run("when second item is invalid then validation fails", func(t *testing.T) {
		order := newValidOrder()
		order.Items = []*OrderItem{
			{Product: &Product{ID: 1}, Quantity: 1},
			{Product: &Product{ID: 2}, Quantity: 0}, // Invalid quantity
			{Product: &Product{ID: 3}, Quantity: 3},
		}

		err := order.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemQuantityInvalid, validationErr.Message)
	})
}

// =============================================================================
// OrderStatus Constants Tests
// =============================================================================

func TestOrderStatus_Constants(t *testing.T) {
	t.Run("draft status has correct value", func(t *testing.T) {
		assert.Equal(t, OrderStatus("draft"), OrderStatusDraft)
	})

	t.Run("pending status has correct value", func(t *testing.T) {
		assert.Equal(t, OrderStatus("pending"), OrderStatusPending)
	})

	t.Run("confirmed status has correct value", func(t *testing.T) {
		assert.Equal(t, OrderStatus("confirmed"), OrderStatusConfirmed)
	})

	t.Run("completed status has correct value", func(t *testing.T) {
		assert.Equal(t, OrderStatus("completed"), OrderStatusCompleted)
	})

	t.Run("cancelled status has correct value", func(t *testing.T) {
		assert.Equal(t, OrderStatus("cancelled"), OrderStatusCancelled)
	})
}

// =============================================================================
// GetID Tests
// =============================================================================

func TestOrder_GetID(t *testing.T) {
	t.Run("when order has ID then returns correct ID", func(t *testing.T) {
		// Arrange
		order := &Order{ID: 42}

		// Act
		result := order.GetID()

		// Assert
		assert.Equal(t, 42, result)
	})

	t.Run("when order has zero ID then returns zero", func(t *testing.T) {
		// Arrange
		order := &Order{}

		// Act
		result := order.GetID()

		// Assert
		assert.Equal(t, 0, result)
	})
}

// =============================================================================
// GetSortValue Tests
// =============================================================================

func TestOrder_GetSortValue(t *testing.T) {
	now := time.Now()

	order := &Order{
		ID:          1,
		OrderNumber: "ORD-001",
		Total:       150.75,
		CreatedAt:   now,
	}

	t.Run("when sortBy is total then returns Total", func(t *testing.T) {
		// Act
		result := order.GetSortValue(OrderSortByTotal)

		// Assert
		assert.Equal(t, 150.75, result)
	})

	t.Run("when sortBy is order_number then returns OrderNumber", func(t *testing.T) {
		// Act
		result := order.GetSortValue(OrderSortByOrderNumber)

		// Assert
		assert.Equal(t, "ORD-001", result)
	})

	t.Run("when sortBy is created_at then returns CreatedAt", func(t *testing.T) {
		// Act
		result := order.GetSortValue(SortByCreatedAt)

		// Assert
		assert.Equal(t, now, result)
	})

	t.Run("when sortBy is unknown then returns nil", func(t *testing.T) {
		// Act
		result := order.GetSortValue("unknown_field")

		// Assert
		assert.Nil(t, result)
	})
}
