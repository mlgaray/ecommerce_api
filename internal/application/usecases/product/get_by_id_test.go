package product

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

func TestGetByIDUseCase_Execute(t *testing.T) {
	t.Run("when product exists then returns product", func(t *testing.T) {
		ctx := context.Background()
		productID := 1
		expected := &models.Product{ID: 1, Name: "Laptop"}

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			GetByID(mock.Anything, productID).
			Return(expected, nil)

		uc := NewGetByIDUseCase(serviceMock)

		result, err := uc.Execute(ctx, productID)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when product not found then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := &domainErrors.RecordNotFoundError{Message: "product_not_found"}

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			GetByID(mock.Anything, 999).
			Return(nil, expectedErr)

		uc := NewGetByIDUseCase(serviceMock)

		result, err := uc.Execute(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.ErrorAs(t, err, &notFound)
	})

	t.Run("when service errors then propagates error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("database failure")

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			GetByID(mock.Anything, 1).
			Return(nil, expectedErr)

		uc := NewGetByIDUseCase(serviceMock)

		result, err := uc.Execute(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
