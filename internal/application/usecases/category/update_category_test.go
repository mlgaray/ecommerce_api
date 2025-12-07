package category

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestUpdateCategoryUseCase_Execute(t *testing.T) {
	t.Run("when update is successful then returns no error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(nil)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when category not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 999
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		var newImageBuffer []byte
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.CategoryNotFound}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(notFoundError)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.CategoryNotFound, notFound.Message)
	})

	t.Run("when duplicate name then returns DuplicateRecordError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Existing Name",
			Description: "Updated Description",
		}
		var newImageBuffer []byte
		duplicateError := &domainErrors.DuplicateRecordError{Message: domainErrors.CategoryAlreadyExistsInShop}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(duplicateError)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		var dupErr *domainErrors.DuplicateRecordError
		assert.True(t, errors.As(err, &dupErr))
		assert.Equal(t, domainErrors.CategoryAlreadyExistsInShop, dupErr.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")
		expectedError := errors.New("service error")

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(expectedError)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when updating without new image then delegates correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
			Image: &models.Image{
				ID:  5,
				URL: "https://cloudinary.com/existing.jpg",
			},
		}
		var newImageBuffer []byte // No new image

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(nil)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when updating with new image then delegates correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			Update(ctx, categoryID, category, newImageBuffer, shopID).
			Return(nil)

		useCase := NewUpdateCategoryUseCase(categoryServiceMock)

		// Act
		err := useCase.Execute(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})
}
