package models

import (
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidOrderItem creates a valid order item for testing
func newValidOrderItem() *OrderItem {
	return &OrderItem{
		Product:  &Product{ID: 1, Name: "Test Product"},
		Quantity: 2,
	}
}

// =============================================================================
// Validate Tests - Product
// =============================================================================

func TestOrderItem_Validate_Product(t *testing.T) {
	t.Run("when product is nil then validation fails", func(t *testing.T) {
		item := newValidOrderItem()
		item.Product = nil

		err := item.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})

	t.Run("when product ID is zero then validation fails", func(t *testing.T) {
		item := newValidOrderItem()
		item.Product = &Product{ID: 0}

		err := item.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})

	t.Run("when product ID is negative then validation passes (negative IDs are technically valid)", func(t *testing.T) {
		// Note: In practice, negative IDs won't exist in the database,
		// but the domain model doesn't enforce this constraint.
		// The database FK constraint will catch this.
		item := newValidOrderItem()
		item.Product = &Product{ID: -1}

		err := item.Validate()

		// Currently passes validation - DB will reject if needed
		assert.NoError(t, err)
	})

	t.Run("when product ID is positive then validation passes", func(t *testing.T) {
		item := newValidOrderItem()
		item.Product = &Product{ID: 1}

		err := item.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Quantity
// =============================================================================

func TestOrderItem_Validate_Quantity(t *testing.T) {
	t.Run("when quantity is zero then validation fails", func(t *testing.T) {
		item := newValidOrderItem()
		item.Quantity = 0

		err := item.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemQuantityInvalid, validationErr.Message)
	})

	t.Run("when quantity is negative then validation fails", func(t *testing.T) {
		item := newValidOrderItem()
		item.Quantity = -1

		err := item.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemQuantityInvalid, validationErr.Message)
	})

	t.Run("when quantity is one then validation passes", func(t *testing.T) {
		item := newValidOrderItem()
		item.Quantity = 1

		err := item.Validate()

		assert.NoError(t, err)
	})

	t.Run("when quantity is large positive number then validation passes", func(t *testing.T) {
		item := newValidOrderItem()
		item.Quantity = 1000

		err := item.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Combined
// =============================================================================

func TestOrderItem_Validate_Combined(t *testing.T) {
	t.Run("when all fields are valid then validation passes", func(t *testing.T) {
		item := &OrderItem{
			ID:         1,
			OrderID:    1,
			Product:    &Product{ID: 1, Name: "Test Product"},
			Quantity:   5,
			UnitPrice:  10.00,
			TotalPrice: 50.00,
		}

		err := item.Validate()

		assert.NoError(t, err)
	})

	t.Run("product is validated before quantity", func(t *testing.T) {
		item := &OrderItem{
			Product:  nil,
			Quantity: 0,
		}

		err := item.Validate()

		// Product error should come first
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})
}
