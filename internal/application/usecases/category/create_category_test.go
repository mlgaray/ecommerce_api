package category

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

func TestCreateCategoryUseCase_Execute(t *testing.T) {
	t.Run("when create is successful then returns category", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		category := &models.Category{Name: "Electronics"}
		imageBuffer := []byte{1, 2, 3}
		expected := &models.Category{ID: 1, Name: "Electronics"}

		serviceMock := mocks.NewCategoryService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, category, imageBuffer, shopID).
			Return(expected, nil)

		uc := NewCreateCategoryUseCase(serviceMock)

		result, err := uc.Execute(ctx, category, imageBuffer, shopID)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when duplicate name then returns error", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := &domainErrors.DuplicateRecordError{Message: "category_name_already_exists"}

		serviceMock := mocks.NewCategoryService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, expectedErr)

		uc := NewCreateCategoryUseCase(serviceMock)

		result, err := uc.Execute(ctx, &models.Category{}, nil, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		var dupErr *domainErrors.DuplicateRecordError
		assert.ErrorAs(t, err, &dupErr)
	})

	t.Run("when service returns unexpected error then propagates", func(t *testing.T) {
		ctx := context.Background()
		expectedErr := errors.New("database error")

		serviceMock := mocks.NewCategoryService(t)
		serviceMock.EXPECT().
			Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, expectedErr)

		uc := NewCreateCategoryUseCase(serviceMock)

		result, err := uc.Execute(ctx, &models.Category{}, nil, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
