package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Order.RemoveCoupon Tests
// =============================================================================

func TestOrder_RemoveCoupon(t *testing.T) {
	t.Run("when order has coupon then removes coupon and recalculates", func(t *testing.T) {
		// Arrange
		order := &Order{
			ID:           1,
			Subtotal:     1000,
			ShippingCost: 150,
			Discount:     100,
			Total:        1050, // 1000 + 150 - 100
			Coupon:       &Coupon{ID: 1, Code: "SUMMER20", Type: CouponTypePercentage, Value: 10},
		}

		// Act
		result := order.RemoveCoupon()

		// Assert
		assert.True(t, result)
		assert.Nil(t, order.Coupon)
		assert.Equal(t, 0.0, order.Discount)
		assert.Equal(t, order.Subtotal+order.ShippingCost, order.Total)
		assert.Equal(t, 1150.0, order.Total)
	})

	t.Run("when order has no coupon then returns false", func(t *testing.T) {
		// Arrange
		order := &Order{
			ID:           2,
			Subtotal:     500,
			ShippingCost: 100,
			Discount:     0,
			Total:        600,
			Coupon:       nil,
		}

		// Act
		result := order.RemoveCoupon()

		// Assert
		assert.False(t, result)
		assert.Nil(t, order.Coupon)
		assert.Equal(t, 0.0, order.Discount)
		assert.Equal(t, 600.0, order.Total)
	})
}
