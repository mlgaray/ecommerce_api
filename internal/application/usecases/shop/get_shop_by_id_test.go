package shop

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidShop creates a valid shop for testing
func newValidShop() *models.Shop {
	return &models.Shop{
		ID:        1,
		Name:      "Test Shop",
		Slug:      "test-shop",
		Email:     "test@shop.com",
		Phone:     "+54 11 1234-5678",
		Instagram: "@testshop",
		Images: []*models.Image{
			{ID: 1, URL: "https://cloudinary.com/logo.jpg", Type: "logo"},
			{ID: 2, URL: "https://cloudinary.com/cover.jpg", Type: "cover"},
		},
		Address: &models.Address{
			ID:      1,
			Name:    "Main Street 123",
			PlaceID: "ChIJ123456789",
			Lat:     -34.6037,
			Lng:     -58.3816,
		},
		PaymentMethods: []*models.PaymentMethod{
			{ID: 1, Name: "Transfer", Code: "transfer", IsActive: true},
		},
		DeliveryMethods: []*models.DeliveryMethod{
			{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
		},
		OperatingSchedules: []*models.OperatingSchedule{
			{ID: 1, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
		},
	}
}

// =============================================================================
// GetShopByIDUseCase Tests
// =============================================================================

func TestGetShopByIDUseCase_Execute(t *testing.T) {
	t.Run("when shop exists then returns shop", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := newValidShop()

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedShop.ID, result.ID)
		assert.Equal(t, expectedShop.Name, result.Name)
		assert.Equal(t, expectedShop.Slug, result.Slug)
		assert.NotNil(t, result.Images)
		assert.Len(t, result.Images, 2)
		assert.NotNil(t, result.Address)
		assert.NotNil(t, result.PaymentMethods)
		assert.NotNil(t, result.DeliveryMethods)
		assert.NotNil(t, result.OperatingSchedules)
	})

	t.Run("when shop not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ShopNotFound}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(nil, notFoundError)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ShopNotFound, notFound.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedError := errors.New("database error")

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(nil, expectedError)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when shop has no images then returns shop without images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:     shopID,
			Name:   "Test Shop",
			Slug:   "test-shop",
			Images: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Images)
	})

	t.Run("when shop has no address then returns shop without address", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:      shopID,
			Name:    "Test Shop",
			Slug:    "test-shop",
			Address: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Address)
	})

	t.Run("when shop has no payment methods then returns shop without payment methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:             shopID,
			Name:           "Test Shop",
			Slug:           "test-shop",
			PaymentMethods: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.PaymentMethods)
	})

	t.Run("when shop has no delivery methods then returns shop without delivery methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:              shopID,
			Name:            "Test Shop",
			Slug:            "test-shop",
			DeliveryMethods: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.DeliveryMethods)
	})

	t.Run("when shop has no operating schedules then returns shop without schedules", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:                 shopID,
			Name:               "Test Shop",
			Slug:               "test-shop",
			OperatingSchedules: nil,
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		useCase := NewGetShopByIDUseCase(repoMock)

		// Act
		result, err := useCase.Execute(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.OperatingSchedules)
	})
}
