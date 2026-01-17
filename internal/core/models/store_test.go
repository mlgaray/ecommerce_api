package models

import (
	"testing"
	"time"

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

	t.Run("when shop has timezone then maps timezone", func(t *testing.T) {
		// Arrange
		shop := &Shop{
			ID:   1,
			Name: "Test Shop",
			Timezone: &Timezone{
				ID:         1,
				Name:       "Buenos Aires",
				Identifier: "America/Buenos_Aires",
				UTCOffset:  "-03:00",
			},
		}

		// Act
		result := NewStoreFromShop(shop)

		// Assert
		assert.NotNil(t, result)
		assert.NotNil(t, result.Timezone)
		assert.Equal(t, "America/Buenos_Aires", result.Timezone.Identifier)
	})
}

// =============================================================================
// CalculateIsOpen Tests
// =============================================================================

func TestStore_CalculateIsOpen(t *testing.T) {
	t.Run("when timezone is nil then returns false", func(t *testing.T) {
		// Arrange
		store := &Store{
			Timezone: nil,
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(time.Now())

		// Assert
		assert.False(t, result)
	})

	t.Run("when timezone identifier is empty then returns false", func(t *testing.T) {
		// Arrange
		store := &Store{
			Timezone: &Timezone{Identifier: ""},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(time.Now())

		// Assert
		assert.False(t, result)
	})

	t.Run("when operating schedules is empty then returns false", func(t *testing.T) {
		// Arrange
		store := &Store{
			Timezone:           &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{},
		}

		// Act
		result := store.CalculateIsOpen(time.Now())

		// Assert
		assert.False(t, result)
	})

	t.Run("when timezone identifier is invalid then returns false", func(t *testing.T) {
		// Arrange
		store := &Store{
			Timezone: &Timezone{Identifier: "Invalid/Timezone"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(time.Now())

		// Assert
		assert.False(t, result)
	})

	t.Run("when current time is within schedule then returns true", func(t *testing.T) {
		// Arrange: Monday at 10:00 in Buenos Aires
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 10, 0, 0, 0, loc) // Jan 8, 2024 is Monday

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.True(t, result)
	})

	t.Run("when current time is before schedule then returns false", func(t *testing.T) {
		// Arrange: Monday at 08:00 in Buenos Aires (before 09:00 opening)
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 8, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.False(t, result)
	})

	t.Run("when current time is after schedule then returns false", func(t *testing.T) {
		// Arrange: Monday at 19:00 in Buenos Aires (after 18:00 closing)
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 19, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.False(t, result)
	})

	t.Run("when day does not match schedule then returns false", func(t *testing.T) {
		// Arrange: Tuesday at 10:00, but schedule is for Monday
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		tuesday := time.Date(2024, 1, 9, 10, 0, 0, 0, loc) // Jan 9, 2024 is Tuesday

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(tuesday)

		// Assert
		assert.False(t, result)
	})

	t.Run("when multiple schedules and one matches then returns true", func(t *testing.T) {
		// Arrange: Wednesday at 15:00
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		wednesday := time.Date(2024, 1, 10, 15, 0, 0, 0, loc) // Jan 10, 2024 is Wednesday

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
				{DayOfWeek: Tuesday, OpenTime: "09:00", CloseTime: "18:00"},
				{DayOfWeek: Wednesday, OpenTime: "10:00", CloseTime: "20:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(wednesday)

		// Assert
		assert.True(t, result)
	})

	t.Run("when split hours and second slot matches then returns true", func(t *testing.T) {
		// Arrange: Monday at 16:00 (second shift: 15:00-19:00)
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 16, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "13:00"}, // Morning
				{DayOfWeek: Monday, OpenTime: "15:00", CloseTime: "19:00"}, // Afternoon
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.True(t, result)
	})

	t.Run("when split hours and between slots then returns false", func(t *testing.T) {
		// Arrange: Monday at 14:00 (between 13:00 and 15:00)
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 14, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "13:00"}, // Morning
				{DayOfWeek: Monday, OpenTime: "15:00", CloseTime: "19:00"}, // Afternoon
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.False(t, result)
	})

	t.Run("when exactly at opening time then returns true", func(t *testing.T) {
		// Arrange: Monday at exactly 09:00
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 9, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.True(t, result)
	})

	t.Run("when exactly at closing time then returns true", func(t *testing.T) {
		// Arrange: Monday at exactly 18:00
		loc, _ := time.LoadLocation("America/Buenos_Aires")
		monday := time.Date(2024, 1, 8, 18, 0, 0, 0, loc)

		store := &Store{
			Timezone: &Timezone{Identifier: "America/Buenos_Aires"},
			OperatingSchedules: []*OperatingSchedule{
				{DayOfWeek: Monday, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		// Act
		result := store.CalculateIsOpen(monday)

		// Assert
		assert.True(t, result)
	})
}
