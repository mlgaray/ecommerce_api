package category

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestDeleteCategoryUseCase_Execute(t *testing.T) {
	t.Run("when delete is successful then returns no error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Delete(ctx, categoryID, shopID).
			Return(nil)

		useCase := NewDeleteCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when category not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 999
		shopID := 1
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.CategoryNotFound}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Delete(ctx, categoryID, shopID).
			Return(notFoundError)

		useCase := NewDeleteCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, shopID)

		// Assert
		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.CategoryNotFound, notFound.Message)
	})

	t.Run("when category has products then returns ReferentialIntegrityError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		fkError := &domainErrors.ReferentialIntegrityError{Message: domainErrors.CategoryHasProducts}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Delete(ctx, categoryID, shopID).
			Return(fkError)

		useCase := NewDeleteCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, shopID)

		// Assert
		assert.Error(t, err)
		var refErr *domainErrors.ReferentialIntegrityError
		assert.True(t, errors.As(err, &refErr))
		assert.Equal(t, domainErrors.CategoryHasProducts, refErr.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		expectedError := errors.New("service error")

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Delete(ctx, categoryID, shopID).
			Return(expectedError)

		useCase := NewDeleteCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
