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

// newValidProduct creates a valid product for testing
func newValidProduct() *Product {
	return &Product{
		ID:          1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       100.00,
		Stock:       10,
		IsActive:    true,
		Category:    &Category{ID: 1, Name: "Test Category"},
	}
}

// =============================================================================
// GetID Tests (Identifiable Interface)
// =============================================================================

func TestProduct_GetID(t *testing.T) {
	t.Run("returns product ID", func(t *testing.T) {
		product := &Product{ID: 42}
		assert.Equal(t, 42, product.GetID())
	})

	t.Run("returns zero when ID not set", func(t *testing.T) {
		product := &Product{}
		assert.Equal(t, 0, product.GetID())
	})
}

// =============================================================================
// GetSortValue Tests (Sortable Interface)
// =============================================================================

func TestProduct_GetSortValue(t *testing.T) {
	now := time.Now()
	product := &Product{
		ID:        1,
		Name:      "Test Product",
		Price:     99.99,
		CreatedAt: now,
	}

	t.Run("returns price when sorting by price", func(t *testing.T) {
		value := product.GetSortValue("price")
		assert.Equal(t, 99.99, value)
	})

	t.Run("returns name when sorting by name", func(t *testing.T) {
		value := product.GetSortValue("name")
		assert.Equal(t, "Test Product", value)
	})

	t.Run("returns created_at when sorting by created_at", func(t *testing.T) {
		value := product.GetSortValue("created_at")
		assert.Equal(t, now, value)
	})

	t.Run("returns nil when sorting by id", func(t *testing.T) {
		value := product.GetSortValue("id")
		assert.Nil(t, value)
	})

	t.Run("returns nil for unsupported sort field", func(t *testing.T) {
		value := product.GetSortValue("unknown_field")
		assert.Nil(t, value)
	})

	t.Run("returns nil for empty sort field", func(t *testing.T) {
		value := product.GetSortValue("")
		assert.Nil(t, value)
	})
}

// =============================================================================
// Validate Tests - Price and Stock
// =============================================================================

func TestProduct_Validate_PriceAndStock(t *testing.T) {
	t.Run("when price is positive then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.Price = 0.01

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when price is zero then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.Price = 0

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("when price is negative then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.Price = -10.00

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("when stock is zero then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.Stock = 0

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when stock is positive then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.Stock = 100

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when stock is negative then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.Stock = -1

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductStockCannotBeNegative, validationErr.Message)
	})
}

// =============================================================================
// Validate Tests - Minimum Stock
// =============================================================================

func TestProduct_Validate_MinimumStock(t *testing.T) {
	t.Run("when minimum stock is zero then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.MinimumStock = 0

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when minimum stock is positive and less than stock then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.Stock = 10
		product.MinimumStock = 5

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when minimum stock equals stock then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.IsStockeable = true
		product.Stock = 10
		product.MinimumStock = 10

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when minimum stock is negative then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.IsStockeable = true
		product.MinimumStock = -1

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductMinimumStockCannotBeNegative, validationErr.Message)
	})

	t.Run("when minimum stock is set but stock is zero then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.IsStockeable = true
		product.Stock = 0
		product.MinimumStock = 5

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.MinimumStockRequiresStock, validationErr.Message)
	})

	t.Run("when minimum stock is greater than stock then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.IsStockeable = true
		product.Stock = 5
		product.MinimumStock = 10

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductMinimumStockCannotBeGreaterThanStock, validationErr.Message)
	})
}

// =============================================================================
// Validate Tests - Promotional Price
// =============================================================================

func TestProduct_Validate_PromotionalPrice(t *testing.T) {
	t.Run("when not promotional then validation passes without promotional price", func(t *testing.T) {
		product := newValidProduct()
		product.IsPromotional = false
		product.PromotionalPrice = 0

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when not promotional with promotional price set then validation passes", func(t *testing.T) {
		// Non-promotional products can have promotional_price set (it's just ignored)
		product := newValidProduct()
		product.IsPromotional = false
		product.PromotionalPrice = 50.00

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when promotional with valid promotional price then validation passes", func(t *testing.T) {
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 75.00 // Lower than regular price

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("when promotional without promotional price then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.IsPromotional = true
		product.PromotionalPrice = 0

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalProductRequiresPromotionalPrice, validationErr.Message)
	})

	t.Run("when promotional with negative promotional price then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.IsPromotional = true
		product.PromotionalPrice = -10.00

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalProductRequiresPromotionalPrice, validationErr.Message)
	})

	t.Run("when promotional price equals regular price then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 100.00 // Equal to regular price

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalPriceMustBeLowerThanRegularPrice, validationErr.Message)
	})

	t.Run("when promotional price is greater than regular price then returns validation error", func(t *testing.T) {
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 150.00 // Greater than regular price

		err := product.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalPriceMustBeLowerThanRegularPrice, validationErr.Message)
	})
}

// =============================================================================
// CanBeSold Tests
// =============================================================================

func TestProduct_CanBeSold(t *testing.T) {
	t.Run("when active and stockeable with stock then can be sold", func(t *testing.T) {
		product := &Product{
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
		}

		assert.True(t, product.CanBeSold())
	})

	t.Run("when active and stockeable but no stock then cannot be sold", func(t *testing.T) {
		product := &Product{
			IsActive:     true,
			IsStockeable: true,
			Stock:        0,
		}

		assert.False(t, product.CanBeSold())
	})

	t.Run("when active and not stockeable then can always be sold", func(t *testing.T) {
		product := &Product{
			IsActive:     true,
			IsStockeable: false,
			Stock:        0,
		}

		assert.True(t, product.CanBeSold())
	})

	t.Run("when inactive with stock then cannot be sold", func(t *testing.T) {
		product := &Product{
			IsActive:     false,
			IsStockeable: true,
			Stock:        10,
		}

		assert.False(t, product.CanBeSold())
	})

	t.Run("when inactive and no stock then cannot be sold", func(t *testing.T) {
		product := &Product{
			IsActive:     false,
			IsStockeable: true,
			Stock:        0,
		}

		assert.False(t, product.CanBeSold())
	})
}

// =============================================================================
// IsLowStock Tests
// =============================================================================

func TestProduct_IsLowStock(t *testing.T) {
	t.Run("when stockeable and stock equals minimum stock then is low stock", func(t *testing.T) {
		product := &Product{
			IsStockeable: true,
			Stock:        5,
			MinimumStock: 5,
		}

		assert.True(t, product.IsLowStock())
	})

	t.Run("when stockeable and stock is below minimum stock then is low stock", func(t *testing.T) {
		product := &Product{
			IsStockeable: true,
			Stock:        3,
			MinimumStock: 5,
		}

		assert.True(t, product.IsLowStock())
	})

	t.Run("when stockeable and stock is above minimum stock then is not low stock", func(t *testing.T) {
		product := &Product{
			IsStockeable: true,
			Stock:        10,
			MinimumStock: 5,
		}

		assert.False(t, product.IsLowStock())
	})

	t.Run("when not stockeable then is never low stock", func(t *testing.T) {
		product := &Product{
			IsStockeable: false,
			Stock:        0,
			MinimumStock: 5,
		}

		assert.False(t, product.IsLowStock())
	})

	t.Run("when stockeable with no minimum stock set and has stock then not low stock", func(t *testing.T) {
		product := &Product{
			IsStockeable: true,
			Stock:        10,
			MinimumStock: 0,
		}

		// Stock (10) <= MinimumStock (0) is false
		assert.False(t, product.IsLowStock())
	})
}

// =============================================================================
// GetEffectivePrice Tests
// =============================================================================

func TestProduct_GetEffectivePrice(t *testing.T) {
	t.Run("when not promotional then returns regular price", func(t *testing.T) {
		product := &Product{
			Price:            100.00,
			IsPromotional:    false,
			PromotionalPrice: 0,
		}

		assert.Equal(t, 100.00, product.GetEffectivePrice())
	})

	t.Run("when not promotional but promotional price set then returns regular price", func(t *testing.T) {
		// Edge case: promotional price set but not active
		product := &Product{
			Price:            100.00,
			IsPromotional:    false,
			PromotionalPrice: 75.00,
		}

		assert.Equal(t, 100.00, product.GetEffectivePrice())
	})

	t.Run("when promotional then returns promotional price", func(t *testing.T) {
		product := &Product{
			Price:            100.00,
			IsPromotional:    true,
			PromotionalPrice: 75.00,
		}

		assert.Equal(t, 75.00, product.GetEffectivePrice())
	})
}

// =============================================================================
// DecrementStock Tests
// =============================================================================

func TestProduct_DecrementStock(t *testing.T) {
	t.Run("when quantity is valid and stock is sufficient then decrements stock", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.DecrementStock(3)

		assert.NoError(t, err)
		assert.Equal(t, 7, product.Stock)
	})

	t.Run("when decrementing entire stock then stock becomes zero", func(t *testing.T) {
		product := &Product{Stock: 5, IsStockeable: true}

		err := product.DecrementStock(5)

		assert.NoError(t, err)
		assert.Equal(t, 0, product.Stock)
	})

	t.Run("when quantity is zero then returns validation error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.DecrementStock(0)

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "quantity_must_be_positive", validationErr.Message)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})

	t.Run("when quantity is negative then returns validation error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.DecrementStock(-5)

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "quantity_must_be_positive", validationErr.Message)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})

	t.Run("when insufficient stock then returns business rule error", func(t *testing.T) {
		product := &Product{Stock: 5, IsStockeable: true}

		err := product.DecrementStock(10)

		assert.Error(t, err)
		var businessErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &businessErr))
		assert.Equal(t, "insufficient_stock", businessErr.Message)
		assert.Equal(t, 5, product.Stock) // Stock unchanged
	})

	t.Run("when stock is zero and trying to decrement then returns business rule error", func(t *testing.T) {
		product := &Product{Stock: 0, IsStockeable: true}

		err := product.DecrementStock(1)

		assert.Error(t, err)
		var businessErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &businessErr))
		assert.Equal(t, "insufficient_stock", businessErr.Message)
	})

	t.Run("when product is not stockeable then does nothing and returns no error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: false}

		err := product.DecrementStock(5)

		assert.NoError(t, err)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})
}

// =============================================================================
// IncrementStock Tests
// =============================================================================

func TestProduct_IncrementStock(t *testing.T) {
	t.Run("when quantity is valid then increments stock", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.IncrementStock(5)

		assert.NoError(t, err)
		assert.Equal(t, 15, product.Stock)
	})

	t.Run("when stock is zero then increments correctly", func(t *testing.T) {
		product := &Product{Stock: 0, IsStockeable: true}

		err := product.IncrementStock(10)

		assert.NoError(t, err)
		assert.Equal(t, 10, product.Stock)
	})

	t.Run("when quantity is zero then returns validation error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.IncrementStock(0)

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "quantity_must_be_positive", validationErr.Message)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})

	t.Run("when quantity is negative then returns validation error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: true}

		err := product.IncrementStock(-5)

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, "quantity_must_be_positive", validationErr.Message)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})

	t.Run("when product is not stockeable then does nothing and returns no error", func(t *testing.T) {
		product := &Product{Stock: 10, IsStockeable: false}

		err := product.IncrementStock(5)

		assert.NoError(t, err)
		assert.Equal(t, 10, product.Stock) // Stock unchanged
	})
}

// =============================================================================
// Edge Cases and Complex Scenarios
// =============================================================================

func TestProduct_EdgeCases(t *testing.T) {
	t.Run("product with all fields at boundary values validates correctly", func(t *testing.T) {
		product := &Product{
			Price:            0.01, // Minimum valid price
			Stock:            0,    // Zero stock is valid
			MinimumStock:     0,    // Zero minimum stock is valid
			IsPromotional:    false,
			PromotionalPrice: 0,
		}

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("promotional product with minimum price difference validates correctly", func(t *testing.T) {
		product := &Product{
			Price:            100.00,
			IsPromotional:    true,
			PromotionalPrice: 99.99, // Just below regular price
			Stock:            10,
		}

		err := product.Validate()

		assert.NoError(t, err)
	})

	t.Run("product with large stock values works correctly", func(t *testing.T) {
		product := &Product{
			Price:        100.00,
			Stock:        1000000,
			MinimumStock: 100000,
		}

		err := product.Validate()

		assert.NoError(t, err)
		assert.False(t, product.IsLowStock())
	})

	t.Run("multiple operations on same product work correctly", func(t *testing.T) {
		product := &Product{
			Price:        100.00,
			Stock:        10,
			IsActive:     true,
			IsStockeable: true,
		}

		// Initial state
		assert.True(t, product.CanBeSold())
		assert.Equal(t, 100.00, product.GetEffectivePrice())

		// Decrement stock
		err := product.DecrementStock(5)
		assert.NoError(t, err)
		assert.Equal(t, 5, product.Stock)
		assert.True(t, product.CanBeSold())

		// Decrement more
		err = product.DecrementStock(5)
		assert.NoError(t, err)
		assert.Equal(t, 0, product.Stock)
		assert.False(t, product.CanBeSold()) // No stock

		// Increment stock
		err = product.IncrementStock(3)
		assert.NoError(t, err)
		assert.Equal(t, 3, product.Stock)
		assert.True(t, product.CanBeSold())
	})

	t.Run("validation order - price error comes before stock error", func(t *testing.T) {
		product := &Product{
			Price: -10.00, // Invalid
			Stock: -5,     // Also invalid
		}

		err := product.Validate()

		// Should get price error first (validation order)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("validation order - stock error comes before minimum stock error", func(t *testing.T) {
		product := &Product{
			Price:        100.00,
			Stock:        -5,  // Invalid
			MinimumStock: -10, // Also invalid
		}

		err := product.Validate()

		// Should get stock error first (validation order)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductStockCannotBeNegative, validationErr.Message)
	})
}

// =============================================================================
// HasSufficientStock Tests
// =============================================================================

func TestProduct_HasSufficientStock(t *testing.T) {
	t.Run("when not stockeable then always returns true", func(t *testing.T) {
		product := &Product{IsStockeable: false, Stock: 0}

		assert.True(t, product.HasSufficientStock(100))
	})

	t.Run("when stockeable and stock equals quantity then returns true", func(t *testing.T) {
		product := &Product{IsStockeable: true, Stock: 5}

		assert.True(t, product.HasSufficientStock(5))
	})

	t.Run("when stockeable and stock greater than quantity then returns true", func(t *testing.T) {
		product := &Product{IsStockeable: true, Stock: 10}

		assert.True(t, product.HasSufficientStock(5))
	})

	t.Run("when stockeable and stock less than quantity then returns false", func(t *testing.T) {
		product := &Product{IsStockeable: true, Stock: 3}

		assert.False(t, product.HasSufficientStock(5))
	})

	t.Run("when stockeable and stock is zero then returns false", func(t *testing.T) {
		product := &Product{IsStockeable: true, Stock: 0}

		assert.False(t, product.HasSufficientStock(1))
	})

	t.Run("when stockeable and quantity is zero then returns true", func(t *testing.T) {
		product := &Product{IsStockeable: true, Stock: 5}

		assert.True(t, product.HasSufficientStock(0))
	})
}

// =============================================================================
// EqualsForOrder Tests
// =============================================================================

func TestProduct_EqualsForOrder(t *testing.T) {
	t.Run("when other is nil then returns false", func(t *testing.T) {
		product := &Product{ID: 1, Name: "Big Mac", Price: 10000}

		assert.False(t, product.EqualsForOrder(nil))
	})

	t.Run("when all fields match then returns true", func(t *testing.T) {
		dbProduct := &Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    false,
			PromotionalPrice: 0,
		}
		frontendProduct := &Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    false,
			PromotionalPrice: 0,
		}

		assert.True(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("when promotional fields match then returns true", func(t *testing.T) {
		dbProduct := &Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    true,
			PromotionalPrice: 8000,
		}
		frontendProduct := &Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    true,
			PromotionalPrice: 8000,
		}

		assert.True(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("when name differs then returns false", func(t *testing.T) {
		dbProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000}
		frontendProduct := &Product{ID: 1, Name: "Wrong Name", Price: 10000}

		assert.False(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("when price differs then returns false", func(t *testing.T) {
		dbProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000}
		frontendProduct := &Product{ID: 1, Name: "Big Mac", Price: 15000}

		assert.False(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("when is_promotional differs then returns false", func(t *testing.T) {
		dbProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000, IsPromotional: false}
		frontendProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000, IsPromotional: true}

		assert.False(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("when promotional_price differs then returns false", func(t *testing.T) {
		dbProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000, IsPromotional: true, PromotionalPrice: 8000}
		frontendProduct := &Product{ID: 1, Name: "Big Mac", Price: 10000, IsPromotional: true, PromotionalPrice: 7000}

		assert.False(t, dbProduct.EqualsForOrder(frontendProduct))
	})

	t.Run("ignores fields not compared (ID, IsActive, Stock, etc)", func(t *testing.T) {
		dbProduct := &Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        100,
		}
		frontendProduct := &Product{
			ID:           999, // Different ID
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     false, // Different
			IsStockeable: false, // Different
			Stock:        0,     // Different
		}

		// Should still match because EqualsForOrder only compares Name, Price, IsPromotional, PromotionalPrice
		assert.True(t, dbProduct.EqualsForOrder(frontendProduct))
	})
}

// =============================================================================
// VariantsEqualForOrder Tests
// =============================================================================

func TestProduct_VariantsEqualForOrder(t *testing.T) {
	t.Run("when no frontend variants then returns true", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}

		assert.True(t, dbProduct.VariantsEqualForOrder([]*Variant{}))
	})

	t.Run("when variant found and options match then returns true", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{
					{ID: 1, Name: "Chico", Price: 0},
					{ID: 2, Name: "Grande", Price: 1500},
				}},
				{ID: 2, Name: "Extras", Options: []*Option{
					{ID: 3, Name: "Extra bacon", Price: 800},
					{ID: 4, Name: "Extra queso", Price: 800},
				}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 2, Name: "Grande", Price: 1500}}},
			{ID: 2, Name: "Extras", Options: []*Option{{ID: 3, Name: "Extra bacon", Price: 800}}},
		}

		assert.True(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when variant not found in db then returns false", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 999, Name: "Unknown", Options: []*Option{{ID: 1, Name: "Option", Price: 0}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when variant name differs then returns false", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Wrong Name", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when option not found in db variant then returns false", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 999, Name: "Unknown", Price: 0}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when option name differs then returns false", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Wrong", Price: 1500}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when option price differs then returns false", func(t *testing.T) {
		dbProduct := &Product{
			Variants: []*Variant{
				{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
			},
		}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 9999}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when db product has no variants then frontend variant not found returns false", func(t *testing.T) {
		dbProduct := &Product{Variants: []*Variant{}}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})

	t.Run("when db product has nil variants then frontend variant not found returns false", func(t *testing.T) {
		dbProduct := &Product{Variants: nil}
		frontendVariants := []*Variant{
			{ID: 1, Name: "Tamaño", Options: []*Option{{ID: 1, Name: "Grande", Price: 1500}}},
		}

		assert.False(t, dbProduct.VariantsEqualForOrder(frontendVariants))
	})
}
