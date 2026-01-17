package services

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
// GetBySlug Tests
// =============================================================================

func TestStoreService_GetBySlug(t *testing.T) {
	t.Run("when store exists then returns store", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-shop"
		expectedShop := &models.Shop{
			ID:    1,
			Name:  "Test Shop",
			Slug:  slug,
			Email: "test@shop.com",
			Images: []*models.Image{
				{ID: 1, URL: "https://cloudinary.com/logo.jpg", Type: "logo"},
			},
			Address: &models.Address{
				ID:   1,
				Name: "Main Street 123",
			},
			PaymentMethods: []*models.PaymentMethod{
				{ID: 1, Name: "Cash", Code: "cash", IsActive: true},
			},
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
			OperatingSchedules: []*models.OperatingSchedule{
				{ID: 1, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedShop.ID, result.ID)
		assert.Equal(t, expectedShop.Name, result.Name)
		assert.Equal(t, expectedShop.Slug, result.Slug)
		assert.Equal(t, expectedShop.Email, result.Email)
		assert.NotNil(t, result.Images)
		assert.NotNil(t, result.Address)
		assert.NotNil(t, result.PaymentMethods)
		assert.NotNil(t, result.DeliveryMethods)
		assert.NotNil(t, result.OperatingSchedules)
	})

	t.Run("when store not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "non-existent-shop"
		notFoundError := &errors.RecordNotFoundError{Message: errors.StoreNotFound}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.StoreNotFound, notFound.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-shop"
		expectedError := stdErrors.New("database error")

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, expectedError)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when store has no images then returns store without images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-images"
		expectedShop := &models.Shop{
			ID:     1,
			Name:   "Shop Without Images",
			Slug:   slug,
			Images: []*models.Image{},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Images)
	})

	t.Run("when store has no address then returns store without address", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-address"
		expectedShop := &models.Shop{
			ID:      1,
			Name:    "Shop Without Address",
			Slug:    slug,
			Address: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Address)
	})

	t.Run("when store has no payment methods then returns store without payment methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-payment-methods"
		expectedShop := &models.Shop{
			ID:             1,
			Name:           "Shop Without Payment Methods",
			Slug:           slug,
			PaymentMethods: []*models.PaymentMethod{},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.PaymentMethods)
	})

	t.Run("when store has no delivery methods then returns store without delivery methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-delivery-methods"
		expectedShop := &models.Shop{
			ID:              1,
			Name:            "Shop Without Delivery Methods",
			Slug:            slug,
			DeliveryMethods: []*models.DeliveryMethod{},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.DeliveryMethods)
	})

	t.Run("when store has no operating schedules then returns store without schedules", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-schedules"
		expectedShop := &models.Shop{
			ID:                 1,
			Name:               "Shop Without Schedules",
			Slug:               slug,
			OperatingSchedules: []*models.OperatingSchedule{},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		service := NewStoreService(repoMock)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.OperatingSchedules)
	})
}
