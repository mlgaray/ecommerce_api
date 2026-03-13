package product

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestCreateProductUseCase_Execute(t *testing.T) {
	t.Run("when create is successful then returns product", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		product := &models.Product{Name: "Test Product"}
		imageBuffers := [][]byte{{1, 2, 3}}
		expected := &models.Product{ID: 1, Name: "Test Product"}

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, product, imageBuffers, shopID).
			Return(expected, nil)

		uc := NewCreateProductUseCase(serviceMock)

		result, err := uc.Execute(ctx, product, imageBuffers, shopID)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns validation error then propagates", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("validation failed")

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, expectedErr)

		uc := NewCreateProductUseCase(serviceMock)

		result, err := uc.Execute(ctx, &models.Product{}, nil, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "validation failed", err.Error())
	})

	t.Run("when service returns unexpected error then propagates", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("database error")

		serviceMock := mocks.NewProductService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, expectedErr)

		uc := NewCreateProductUseCase(serviceMock)

		result, err := uc.Execute(ctx, &models.Product{}, nil, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
