package responses

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestShop() *models.Shop {
	primaryColor := "#FF5733"
	fixedPrice := 500.0
	return &models.Shop{
		ID:           1,
		Name:         "Mi Tienda",
		Slug:         "mi-tienda",
		Email:        "shop@example.com",
		Phone:        "+54 11 1234-5678",
		Instagram:    "@mitienda",
		PrimaryColor: &primaryColor,
		Images: []*models.Image{
			{ID: 10, URL: "https://example.com/logo.jpg", Type: "logo"},
			{ID: 11, URL: "https://example.com/cover.jpg", Type: "cover"},
		},
		Address: &models.Address{
			ID:      5,
			Name:    "Av. Corrientes 1234",
			PlaceID: "ChIJxyz",
			Lat:     -34.6037,
			Lng:     -58.3816,
		},
		PaymentMethods: []*models.PaymentMethod{
			{
				ID:       1,
				Name:     "Transfer",
				Code:     models.PaymentMethodTransfer,
				IsActive: true,
				TransferConfig: &models.TransferConfig{
					ID:        1,
					CBU:       "0123456789012345678901",
					CUIL:      "20-12345678-9",
					Alias:     "mi.tienda",
					OwnerName: "Juan Perez",
				},
			},
			{
				ID:       2,
				Name:     "MercadoPago",
				Code:     models.PaymentMethodMercadoPago,
				IsActive: true,
				MercadoPagoConfig: &models.MercadoPagoConfig{
					ID:          1,
					AccessToken: "TEST-TOKEN",
					PublicKey:   "TEST-KEY",
					UserID:      "12345",
				},
			},
		},
		DeliveryMethods: []*models.DeliveryMethod{
			{
				ID:       1,
				Name:     "Delivery",
				Code:     models.DeliveryMethodDelivery,
				IsActive: true,
				DeliveryConfig: &models.DeliveryConfig{
					ID:         1,
					FixedPrice: &fixedPrice,
				},
				DeliveryZones: []*models.DeliveryZone{
					{ID: 1, Name: "Zona Norte", Price: 300},
					{ID: 2, Name: "Zona Sur", Price: 400},
				},
			},
			{
				ID:       2,
				Name:     "Pickup",
				Code:     models.DeliveryMethodPickup,
				IsActive: true,
				PickupConfig: &models.PickupConfig{
					ID:           1,
					Address:      "Av. Corrientes 1234",
					City:         "CABA",
					Province:     "Buenos Aires",
					PostalCode:   "C1043",
					Instructions: "Tocar timbre 4B",
				},
			},
		},
		OperatingSchedules: []*models.OperatingSchedule{
			{ID: 1, DayOfWeek: models.Monday, OpenTime: "09:00", CloseTime: "18:00"},
			{ID: 2, DayOfWeek: models.Tuesday, OpenTime: "09:00", CloseTime: "18:00"},
		},
		Timezone: &models.Timezone{
			ID:         1,
			Name:       "Argentina",
			Identifier: "America/Argentina/Buenos_Aires",
			UTCOffset:  "-03:00",
		},
	}
}

// =============================================================================
// ShopResponseFromModel Tests
// =============================================================================

func TestShopResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when full shop then maps all fields", func(t *testing.T) {
		// Arrange
		shop := newTestShop()

		// Act
		result := ShopResponseFromModel(shop)

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
		assert.Len(t, result.Images, 2)
		assert.NotNil(t, result.Address)
		assert.Len(t, result.PaymentMethods, 2)
		assert.Len(t, result.DeliveryMethods, 2)
		assert.Len(t, result.OperatingSchedules, 2)
		assert.NotNil(t, result.Timezone)
	})

	t.Run("when shop with nil optional fields then maps nil", func(t *testing.T) {
		// Arrange
		shop := &models.Shop{
			ID:   2,
			Name: "Minimal Shop",
			Slug: "minimal",
		}

		// Act
		result := ShopResponseFromModel(shop)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.ID)
		assert.Nil(t, result.PrimaryColor)
		assert.Nil(t, result.Images)
		assert.Nil(t, result.Address)
		assert.Nil(t, result.PaymentMethods)
		assert.Nil(t, result.DeliveryMethods)
		assert.Nil(t, result.OperatingSchedules)
		assert.Nil(t, result.Timezone)
	})
}

// =============================================================================
// ShopImageResponseFromModel Tests
// =============================================================================

func TestShopImageResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopImageResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when image then maps all fields", func(t *testing.T) {
		// Arrange
		img := &models.Image{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"}

		// Act
		result := ShopImageResponseFromModel(img)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "https://example.com/logo.jpg", result.URL)
		assert.Equal(t, "logo", result.Type)
	})
}

func TestShopImageResponsesFromModels(t *testing.T) {
	t.Run("when empty then returns nil", func(t *testing.T) {
		// Act
		result := shopImageResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when images then maps all", func(t *testing.T) {
		// Arrange
		images := []*models.Image{
			{ID: 1, URL: "https://example.com/a.jpg", Type: "logo"},
			{ID: 2, URL: "https://example.com/b.jpg", Type: "cover"},
		}

		// Act
		result := shopImageResponsesFromModels(images)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "logo", result[0].Type)
		assert.Equal(t, "cover", result[1].Type)
	})
}

// =============================================================================
// ShopAddressResponseFromModel Tests
// =============================================================================

func TestShopAddressResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopAddressResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when address then maps all fields", func(t *testing.T) {
		// Arrange
		address := &models.Address{
			ID:      5,
			Name:    "Av. Corrientes 1234",
			PlaceID: "ChIJxyz",
			Lat:     -34.6037,
			Lng:     -58.3816,
		}

		// Act
		result := ShopAddressResponseFromModel(address)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 5, result.ID)
		assert.Equal(t, "Av. Corrientes 1234", result.Name)
		assert.Equal(t, "ChIJxyz", result.PlaceID)
		assert.Equal(t, -34.6037, result.Lat)
		assert.Equal(t, -58.3816, result.Lng)
	})
}

// =============================================================================
// ShopPaymentMethodResponseFromModel Tests
// =============================================================================

func TestShopPaymentMethodResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopPaymentMethodResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when payment method with transfer config then maps all", func(t *testing.T) {
		// Arrange
		pm := &models.PaymentMethod{
			ID:       1,
			Name:     "Transfer",
			Code:     models.PaymentMethodTransfer,
			IsActive: true,
			TransferConfig: &models.TransferConfig{
				ID:        1,
				CBU:       "0123456789012345678901",
				CUIL:      "20-12345678-9",
				Alias:     "mi.tienda",
				OwnerName: "Juan",
			},
		}

		// Act
		result := ShopPaymentMethodResponseFromModel(pm)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Transfer", result.Name)
		assert.Equal(t, "transfer", result.Code)
		assert.True(t, result.IsActive)
		assert.NotNil(t, result.TransferConfig)
		assert.Equal(t, "0123456789012345678901", result.TransferConfig.CBU)
		assert.Equal(t, "mi.tienda", result.TransferConfig.Alias)
		assert.Nil(t, result.MercadoPagoConfig)
	})

	t.Run("when payment method with mercadopago config then maps all", func(t *testing.T) {
		// Arrange
		pm := &models.PaymentMethod{
			ID:       2,
			Name:     "MercadoPago",
			Code:     models.PaymentMethodMercadoPago,
			IsActive: true,
			MercadoPagoConfig: &models.MercadoPagoConfig{
				ID:          1,
				AccessToken: "TOKEN",
				PublicKey:   "KEY",
				UserID:      "12345",
			},
		}

		// Act
		result := ShopPaymentMethodResponseFromModel(pm)

		// Assert
		assert.NotNil(t, result.MercadoPagoConfig)
		assert.Equal(t, "TOKEN", result.MercadoPagoConfig.AccessToken)
		assert.Equal(t, "KEY", result.MercadoPagoConfig.PublicKey)
		assert.Equal(t, "12345", result.MercadoPagoConfig.UserID)
		assert.Nil(t, result.TransferConfig)
	})

	t.Run("when payment method with no configs then configs are nil", func(t *testing.T) {
		// Arrange
		pm := &models.PaymentMethod{
			ID:   3,
			Name: "Cash",
			Code: models.PaymentMethodCash,
		}

		// Act
		result := ShopPaymentMethodResponseFromModel(pm)

		// Assert
		assert.Nil(t, result.TransferConfig)
		assert.Nil(t, result.MercadoPagoConfig)
	})
}

func TestShopPaymentMethodResponsesFromModels(t *testing.T) {
	t.Run("when empty then returns nil", func(t *testing.T) {
		// Act
		result := shopPaymentMethodResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when methods then maps all", func(t *testing.T) {
		// Arrange
		methods := []*models.PaymentMethod{
			{ID: 1, Name: "Cash", Code: models.PaymentMethodCash},
			{ID: 2, Name: "Transfer", Code: models.PaymentMethodTransfer},
		}

		// Act
		result := shopPaymentMethodResponsesFromModels(methods)

		// Assert
		assert.Len(t, result, 2)
	})
}

// =============================================================================
// ShopTransferConfigResponseFromModel Tests
// =============================================================================

func TestShopTransferConfigResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopTransferConfigResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when config then maps all fields", func(t *testing.T) {
		// Arrange
		config := &models.TransferConfig{
			ID:        1,
			CBU:       "0123456789012345678901",
			CUIL:      "20-12345678-9",
			Alias:     "mi.alias",
			OwnerName: "Owner",
		}

		// Act
		result := ShopTransferConfigResponseFromModel(config)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "0123456789012345678901", result.CBU)
		assert.Equal(t, "20-12345678-9", result.CUIL)
		assert.Equal(t, "mi.alias", result.Alias)
		assert.Equal(t, "Owner", result.OwnerName)
	})
}

// =============================================================================
// ShopMercadoPagoConfigResponseFromModel Tests
// =============================================================================

func TestShopMercadoPagoConfigResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopMercadoPagoConfigResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when config then maps all fields", func(t *testing.T) {
		// Arrange
		config := &models.MercadoPagoConfig{
			ID:          1,
			AccessToken: "ACCESS",
			PublicKey:   "PUBLIC",
			UserID:      "UID",
		}

		// Act
		result := ShopMercadoPagoConfigResponseFromModel(config)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "ACCESS", result.AccessToken)
		assert.Equal(t, "PUBLIC", result.PublicKey)
		assert.Equal(t, "UID", result.UserID)
	})
}

// =============================================================================
// ShopDeliveryMethodResponseFromModel Tests
// =============================================================================

func TestShopDeliveryMethodResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopDeliveryMethodResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when delivery method with config and zones then maps all", func(t *testing.T) {
		// Arrange
		fixedPrice := 500.0
		dm := &models.DeliveryMethod{
			ID:       1,
			Name:     "Delivery",
			Code:     models.DeliveryMethodDelivery,
			IsActive: true,
			DeliveryConfig: &models.DeliveryConfig{
				ID:         1,
				FixedPrice: &fixedPrice,
			},
			DeliveryZones: []*models.DeliveryZone{
				{ID: 1, Name: "Zona Norte", Price: 300},
			},
		}

		// Act
		result := ShopDeliveryMethodResponseFromModel(dm)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Delivery", result.Name)
		assert.Equal(t, "delivery", result.Code)
		assert.True(t, result.IsActive)
		assert.NotNil(t, result.DeliveryConfig)
		assert.Equal(t, 500.0, *result.DeliveryConfig.FixedPrice)
		assert.Len(t, result.DeliveryZones, 1)
		assert.Nil(t, result.PickupConfig)
	})

	t.Run("when delivery method with pickup config then maps pickup", func(t *testing.T) {
		// Arrange
		dm := &models.DeliveryMethod{
			ID:       2,
			Name:     "Pickup",
			Code:     models.DeliveryMethodPickup,
			IsActive: true,
			PickupConfig: &models.PickupConfig{
				ID:           1,
				Address:      "Av. Corrientes",
				City:         "CABA",
				Province:     "Buenos Aires",
				PostalCode:   "C1043",
				Instructions: "Ring bell",
			},
		}

		// Act
		result := ShopDeliveryMethodResponseFromModel(dm)

		// Assert
		assert.NotNil(t, result.PickupConfig)
		assert.Equal(t, "Av. Corrientes", result.PickupConfig.Address)
		assert.Equal(t, "CABA", result.PickupConfig.City)
		assert.Equal(t, "Buenos Aires", result.PickupConfig.Province)
		assert.Equal(t, "C1043", result.PickupConfig.PostalCode)
		assert.Equal(t, "Ring bell", result.PickupConfig.Instructions)
		assert.Nil(t, result.DeliveryConfig)
		assert.Nil(t, result.DeliveryZones)
	})
}

func TestShopDeliveryMethodResponsesFromModels(t *testing.T) {
	t.Run("when empty then returns nil", func(t *testing.T) {
		// Act
		result := shopDeliveryMethodResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when methods then maps all", func(t *testing.T) {
		// Arrange
		methods := []*models.DeliveryMethod{
			{ID: 1, Name: "Delivery", Code: models.DeliveryMethodDelivery},
			{ID: 2, Name: "Pickup", Code: models.DeliveryMethodPickup},
		}

		// Act
		result := shopDeliveryMethodResponsesFromModels(methods)

		// Assert
		assert.Len(t, result, 2)
	})
}

// =============================================================================
// ShopDeliveryConfigResponseFromModel Tests
// =============================================================================

func TestShopDeliveryConfigResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopDeliveryConfigResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when config with fixed price then maps all", func(t *testing.T) {
		// Arrange
		fixedPrice := 250.0
		config := &models.DeliveryConfig{ID: 1, FixedPrice: &fixedPrice}

		// Act
		result := ShopDeliveryConfigResponseFromModel(config)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.NotNil(t, result.FixedPrice)
		assert.Equal(t, 250.0, *result.FixedPrice)
	})

	t.Run("when config without fixed price then fixed_price is nil", func(t *testing.T) {
		// Arrange
		config := &models.DeliveryConfig{ID: 1}

		// Act
		result := ShopDeliveryConfigResponseFromModel(config)

		// Assert
		assert.Nil(t, result.FixedPrice)
	})
}

// =============================================================================
// ShopDeliveryZoneResponseFromModel Tests
// =============================================================================

func TestShopDeliveryZoneResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopDeliveryZoneResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when zone then maps all fields", func(t *testing.T) {
		// Arrange
		zone := &models.DeliveryZone{ID: 1, Name: "Zona Norte", Price: 300}

		// Act
		result := ShopDeliveryZoneResponseFromModel(zone)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Zona Norte", result.Name)
		assert.Equal(t, 300.0, result.Price)
	})
}

func TestShopDeliveryZoneResponsesFromModels(t *testing.T) {
	t.Run("when empty then returns nil", func(t *testing.T) {
		// Act
		result := shopDeliveryZoneResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when zones then maps all", func(t *testing.T) {
		// Arrange
		zones := []*models.DeliveryZone{
			{ID: 1, Name: "Norte", Price: 300},
			{ID: 2, Name: "Sur", Price: 400},
		}

		// Act
		result := shopDeliveryZoneResponsesFromModels(zones)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "Norte", result[0].Name)
		assert.Equal(t, "Sur", result[1].Name)
	})
}

// =============================================================================
// ShopPickupConfigResponseFromModel Tests
// =============================================================================

func TestShopPickupConfigResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopPickupConfigResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when config then maps all fields", func(t *testing.T) {
		// Arrange
		config := &models.PickupConfig{
			ID:           1,
			Address:      "Calle 123",
			City:         "CABA",
			Province:     "Buenos Aires",
			PostalCode:   "C1000",
			Instructions: "Puerta roja",
		}

		// Act
		result := ShopPickupConfigResponseFromModel(config)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Calle 123", result.Address)
		assert.Equal(t, "CABA", result.City)
		assert.Equal(t, "Buenos Aires", result.Province)
		assert.Equal(t, "C1000", result.PostalCode)
		assert.Equal(t, "Puerta roja", result.Instructions)
	})
}

// =============================================================================
// ShopOperatingScheduleResponseFromModel Tests
// =============================================================================

func TestShopOperatingScheduleResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopOperatingScheduleResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when schedule then maps all fields", func(t *testing.T) {
		// Arrange
		schedule := &models.OperatingSchedule{
			ID:        1,
			DayOfWeek: models.Monday,
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}

		// Act
		result := ShopOperatingScheduleResponseFromModel(schedule)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, 1, result.DayOfWeek)
		assert.Equal(t, "09:00", result.OpenTime)
		assert.Equal(t, "18:00", result.CloseTime)
	})

	t.Run("when sunday then day_of_week is 0", func(t *testing.T) {
		// Arrange
		schedule := &models.OperatingSchedule{
			DayOfWeek: models.Sunday,
			OpenTime:  "10:00",
			CloseTime: "14:00",
		}

		// Act
		result := ShopOperatingScheduleResponseFromModel(schedule)

		// Assert
		assert.Equal(t, 0, result.DayOfWeek)
	})
}

func TestShopOperatingScheduleResponsesFromModels(t *testing.T) {
	t.Run("when empty then returns nil", func(t *testing.T) {
		// Act
		result := shopOperatingScheduleResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when schedules then maps all", func(t *testing.T) {
		// Arrange
		schedules := []*models.OperatingSchedule{
			{ID: 1, DayOfWeek: models.Monday, OpenTime: "09:00", CloseTime: "18:00"},
			{ID: 2, DayOfWeek: models.Friday, OpenTime: "09:00", CloseTime: "20:00"},
		}

		// Act
		result := shopOperatingScheduleResponsesFromModels(schedules)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].DayOfWeek) // Monday
		assert.Equal(t, 5, result[1].DayOfWeek) // Friday
	})
}

// =============================================================================
// ShopTimezoneResponseFromModel Tests
// =============================================================================

func TestShopTimezoneResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ShopTimezoneResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when timezone then maps all fields", func(t *testing.T) {
		// Arrange
		tz := &models.Timezone{
			ID:         1,
			Name:       "Argentina",
			Identifier: "America/Argentina/Buenos_Aires",
			UTCOffset:  "-03:00",
		}

		// Act
		result := ShopTimezoneResponseFromModel(tz)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Argentina", result.Name)
		assert.Equal(t, "America/Argentina/Buenos_Aires", result.Identifier)
		assert.Equal(t, "-03:00", result.UTCOffset)
	})
}
