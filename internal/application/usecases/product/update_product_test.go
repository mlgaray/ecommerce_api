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

func TestUpdateProductUseCase_Execute(t *testing.T) {
	t.Run("when update is successful then returns no error", func(t *testing.T) {
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := &models.Product{Name: "Updated Product"}
		imageBuffers := [][]byte{{1, 2, 3}}

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Update(mock.Anything, productID, product, imageBuffers, shopID).
			Return(nil)

		uc := NewUpdateProductUseCase(serviceMock)

		err := uc.Execute(ctx, productID, product, imageBuffers, shopID)

		assert.NoError(t, err)
	})

	t.Run("when product not found then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := &domainErrors.RecordNotFoundError{Message: "product_not_found"}

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(expectedErr)

		uc := NewUpdateProductUseCase(serviceMock)

		err := uc.Execute(ctx, 999, &models.Product{}, nil, 1)

		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.ErrorAs(t, err, &notFound)
	})

	t.Run("when service returns unexpected error then propagates", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("database error")

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Update(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(expectedErr)

		uc := NewUpdateProductUseCase(serviceMock)

		err := uc.Execute(ctx, 1, &models.Product{}, nil, 1)

		assert.Error(t, err)
	})
}
