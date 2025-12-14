package product

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestDeleteProductUseCase_Execute(t *testing.T) {
	t.Run("when delete is successful then returns no error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			Delete(ctx, productID, shopID).
			Return(nil)

		useCase := NewDeleteProductUseCase(productServiceMock)

		// Act
		err := useCase.Execute(ctx, productID, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when product not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 999
		shopID := 1
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ProductNotFound}

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			Delete(ctx, productID, shopID).
			Return(notFoundError)

		useCase := NewDeleteProductUseCase(productServiceMock)

		// Act
		err := useCase.Execute(ctx, productID, shopID)

		// Assert
		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ProductNotFound, notFound.Message)
	})

	t.Run("when product belongs to different shop then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 2 // Different shop
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ProductNotFound}

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			Delete(ctx, productID, shopID).
			Return(notFoundError)

		useCase := NewDeleteProductUseCase(productServiceMock)

		// Act
		err := useCase.Execute(ctx, productID, shopID)

		// Assert
		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ProductNotFound, notFound.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		expectedError := errors.New("database operation failed")

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			Delete(ctx, productID, shopID).
			Return(expectedError)

		useCase := NewDeleteProductUseCase(productServiceMock)

		// Act
		err := useCase.Execute(ctx, productID, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
