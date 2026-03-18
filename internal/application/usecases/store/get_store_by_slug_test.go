package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidStore creates a valid store for testing
func newValidStore() *models.Store {
	return &models.Store{
		ID:        1,
		Name:      "Test Store",
		Slug:      "test-store",
		Email:     "test@store.com",
		Phone:     "+54 11 1234-5678",
		Instagram: "@teststore",
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
// GetStoreBySlugUseCase Tests
// =============================================================================

func TestGetStoreBySlugUseCase_Execute(t *testing.T) {
	t.Run("when store exists then returns store", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		expectedStore := newValidStore()

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)
		serviceMock.EXPECT().
			RecordVisit(mock.Anything, expectedStore.ID).
			Return(nil).Maybe()

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedStore.ID, result.ID)
		assert.Equal(t, expectedStore.Name, result.Name)
		assert.Equal(t, expectedStore.Slug, result.Slug)
		assert.NotNil(t, result.Images)
		assert.Len(t, result.Images, 2)
		assert.NotNil(t, result.Address)
		assert.NotNil(t, result.PaymentMethods)
		assert.NotNil(t, result.DeliveryMethods)
		assert.NotNil(t, result.OperatingSchedules)
	})

	t.Run("when store not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "non-existent-store"
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound}

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.StoreNotFound, notFound.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		expectedError := errors.New("database error")

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, expectedError)

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when store has no images then returns store without images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "store-without-images"
		expectedStore := &models.Store{
			ID:     1,
			Name:   "Test Store",
			Slug:   slug,
			Images: nil,
		}

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)
		serviceMock.EXPECT().
			RecordVisit(mock.Anything, expectedStore.ID).
			Return(nil).Maybe()

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Images)
	})

	t.Run("when store has no address then returns store without address", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "store-without-address"
		expectedStore := &models.Store{
			ID:      1,
			Name:    "Test Store",
			Slug:    slug,
			Address: nil,
		}

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)
		serviceMock.EXPECT().
			RecordVisit(mock.Anything, expectedStore.ID).
			Return(nil).Maybe()

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Address)
	})

	t.Run("when store has no operating schedules then returns store without schedules", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "store-without-schedules"
		expectedStore := &models.Store{
			ID:                 1,
			Name:               "Test Store",
			Slug:               slug,
			OperatingSchedules: nil,
		}

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)
		serviceMock.EXPECT().
			RecordVisit(mock.Anything, expectedStore.ID).
			Return(nil).Maybe()

		useCase := NewGetStoreBySlugUseCase(serviceMock)

		// Act
		result, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.OperatingSchedules)
	})
}
