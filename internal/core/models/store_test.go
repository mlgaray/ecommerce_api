package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// NewStoreFromShop Tests
// =============================================================================

func TestNewStoreFromShop(t *testing.T) {
	t.Run("when shop is nil then returns nil", func(t *testing.T) {
		// Act
		result := NewStoreFromShop(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when shop has all fields then maps all fields correctly", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:        1,
			Name:      "Test Shop",
			Slug:      "test-shop",
			Email:     "test@shop.com",
			Phone:     "+54 11 1234-5678",
			Instagram: "@testshop",
			Images: []*Image{
				{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
				{ID: 2, URL: "https://example.com/cover.jpg", Type: "cover"},
			},
			Address: &Address{
				ID:   1,
				Name: "Main Street 123",
			},
			PaymentMethods: []*PaymentMethod{
				{ID: 1, Name: "Cash", Code: "cash", IsActive: true},
			},
			DeliveryMethods: []*DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
			OperatingSchedules: []*OperatingSchedule{
				{ID: 1, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, shop.ID, result.ID)
		assert.Equal(t, shop.Name, result.Name)
		assert.Equal(t, shop.Slug, result.Slug)
		assert.Equal(t, shop.Email, result.Email)
		assert.Equal(t, shop.Phone, result.Phone)
		assert.Equal(t, shop.Instagram, result.Instagram)
		assert.Equal(t, shop.Images, result.Images)
		assert.Equal(t, shop.Address, result.Address)
		assert.Equal(t, shop.PaymentMethods, result.PaymentMethods)
		assert.Equal(t, shop.DeliveryMethods, result.DeliveryMethods)
		assert.Equal(t, shop.OperatingSchedules, result.OperatingSchedules)
	})

	t.Run("when shop has empty name then maps empty name", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:   1,
			Name: "",
			Slug: "test-shop",
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Name)
	})

	t.Run("when shop has empty slug then maps empty slug", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:   1,
			Name: "Test Shop",
			Slug: "",
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, "", result.Slug)
	})

	t.Run("when shop has nil slices then maps nil slices", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:                 1,
			Name:               "Test Shop",
			Images:             nil,
			PaymentMethods:     nil,
			DeliveryMethods:    nil,
			OperatingSchedules: nil,
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Nil(t, result.Images)
		assert.Nil(t, result.PaymentMethods)
		assert.Nil(t, result.DeliveryMethods)
		assert.Nil(t, result.OperatingSchedules)
	})

	t.Run("when shop has empty slices then maps empty slices", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:                 1,
			Name:               "Test Shop",
			Images:             []*Image{},
			PaymentMethods:     []*PaymentMethod{},
			DeliveryMethods:    []*DeliveryMethod{},
			OperatingSchedules: []*OperatingSchedule{},
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.NotNil(t, result.Images)
		assert.Empty(t, result.Images)
		assert.NotNil(t, result.PaymentMethods)
		assert.Empty(t, result.PaymentMethods)
		assert.NotNil(t, result.DeliveryMethods)
		assert.Empty(t, result.DeliveryMethods)
		assert.NotNil(t, result.OperatingSchedules)
		assert.Empty(t, result.OperatingSchedules)
	})

	t.Run("when shop has multiple images then maps all images", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:   1,
			Name: "Test Shop",
			Images: []*Image{
				{ID: 1, URL: "https://example.com/1.jpg", Type: "logo"},
				{ID: 2, URL: "https://example.com/2.jpg", Type: "cover"},
				{ID: 3, URL: "https://example.com/3.jpg", Type: "banner"},
				{ID: 4, URL: "https://example.com/4.jpg", Type: "gallery"},
				{ID: 5, URL: "https://example.com/5.jpg", Type: "gallery"},
			},
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result.Images, 5)
	})

	t.Run("when shop has nil address then maps nil address", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:      1,
			Name:    "Test Shop",
			Address: nil,
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Nil(t, result.Address)
	})

	t.Run("when shop has zero ID then maps zero ID", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:   0,
			Name: "Test Shop",
			Slug: "test-shop",
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ID)
	})
}
