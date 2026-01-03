package shop

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
// UpdateShopUseCase Tests
// =============================================================================

func TestUpdateShopUseCase_Execute(t *testing.T) {
	t.Run("when update is successful with images then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name:  "Updated Shop",
			Slug:  "updated-shop",
			Email: "updated@shop.com",
		}
		logoBuffer := []byte("logo_data")
		coverBuffer := []byte("cover_data")

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, logoBuffer, coverBuffer).
			Return(nil)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when update is successful without images then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name:  "Updated Shop",
			Slug:  "updated-shop",
			Email: "updated@shop.com",
		}
		var logoBuffer []byte
		var coverBuffer []byte

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, logoBuffer, coverBuffer).
			Return(nil)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when update is successful with logo only then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
		}
		logoBuffer := []byte("logo_data")
		var coverBuffer []byte

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, logoBuffer, coverBuffer).
			Return(nil)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when update is successful with cover only then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
		}
		var logoBuffer []byte
		coverBuffer := []byte("cover_data")

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, logoBuffer, coverBuffer).
			Return(nil)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when shop not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
		}
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ShopNotFound}

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, mock.Anything, mock.Anything).
			Return(notFoundError)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, nil, nil)

		// Assert
		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ShopNotFound, notFound.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
		}
		expectedError := errors.New("service error")

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, mock.Anything, mock.Anything).
			Return(expectedError)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, nil, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when validation error then returns ValidationError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
			PaymentMethods: []*models.PaymentMethod{
				{
					ID:       1,
					Code:     "transfer",
					IsActive: true,
					TransferConfig: &models.TransferConfig{
						CBU: "invalid", // Invalid CBU
					},
				},
			},
		}
		validationError := &domainErrors.ValidationError{Message: domainErrors.TransferCBUInvalidLength}

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, mock.Anything, mock.Anything).
			Return(validationError)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, nil, nil)

		// Assert
		assert.Error(t, err)
		var validErr *domainErrors.ValidationError
		assert.True(t, errors.As(err, &validErr))
		assert.Equal(t, domainErrors.TransferCBUInvalidLength, validErr.Message)
	})

	t.Run("when upload fails then returns upload error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := &models.Shop{
			Name: "Updated Shop",
			Slug: "updated-shop",
		}
		logoBuffer := []byte("logo_data")
		uploadError := errors.New("upload failed")

		serviceMock := mocks.NewShopService(t)
		serviceMock.EXPECT().
			Update(ctx, shopID, shop, logoBuffer, mock.Anything).
			Return(uploadError)

		useCase := NewUpdateShopUseCase(serviceMock)

		// Act
		err := useCase.Execute(ctx, shopID, shop, logoBuffer, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uploadError, err)
	})
}
