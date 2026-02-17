package responses

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// =============================================================================
// StoreResponseFromModel Tests
// =============================================================================

func TestStoreResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := StoreResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when full store then maps all fields including IsOpen", func(t *testing.T) {
		// Arrange
		primaryColor := "#FF5733"
		store := &models.Store{
			ID:           1,
			Name:         "Mi Tienda",
			Slug:         "mi-tienda",
			Email:        "shop@example.com",
			Phone:        "+54 11 1234-5678",
			Instagram:    "@mitienda",
			PrimaryColor: &primaryColor,
			Images: []*models.Image{
				{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
			},
			Address: &models.Address{
				ID:   1,
				Name: "Av. Corrientes 1234",
			},
			PaymentMethods: []*models.PaymentMethod{
				{ID: 1, Name: "Cash", Code: models.PaymentMethodCash},
			},
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: models.DeliveryMethodDelivery},
			},
			OperatingSchedules: []*models.OperatingSchedule{
				{ID: 1, DayOfWeek: models.Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
			Timezone: &models.Timezone{
				ID:         1,
				Name:       "Argentina",
				Identifier: "America/Argentina/Buenos_Aires",
				UTCOffset:  "-03:00",
			},
			IsOpen: true,
		}

		// Act
		result := StoreResponseFromModel(store)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Mi Tienda", result.Name)
		assert.Equal(t, "mi-tienda", result.Slug)
		assert.Equal(t, "shop@example.com", result.Email)
		assert.Equal(t, "+54 11 1234-5678", result.Phone)
		assert.Equal(t, "@mitienda", result.Instagram)
		assert.NotNil(t, result.PrimaryColor)
		assert.Equal(t, "#FF5733", *result.PrimaryColor)
		assert.True(t, result.IsOpen)
		assert.Len(t, result.Images, 1)
		assert.NotNil(t, result.Address)
		assert.Len(t, result.PaymentMethods, 1)
		assert.Len(t, result.DeliveryMethods, 1)
		assert.Len(t, result.OperatingSchedules, 1)
		assert.NotNil(t, result.Timezone)
	})

	t.Run("when store is closed then IsOpen is false", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			ID:     2,
			Name:   "Closed Store",
			Slug:   "closed",
			IsOpen: false,
		}

		// Act
		result := StoreResponseFromModel(store)

		// Assert
		assert.NotNil(t, result)
		assert.False(t, result.IsOpen)
	})

	t.Run("when store with nil optional fields then maps nil", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			ID:   3,
			Name: "Minimal",
			Slug: "minimal",
		}

		// Act
		result := StoreResponseFromModel(store)

		// Assert
		assert.Nil(t, result.PrimaryColor)
		assert.Nil(t, result.Images)
		assert.Nil(t, result.Address)
		assert.Nil(t, result.PaymentMethods)
		assert.Nil(t, result.DeliveryMethods)
		assert.Nil(t, result.OperatingSchedules)
		assert.Nil(t, result.Timezone)
	})
}
