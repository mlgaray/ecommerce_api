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

func TestGetByIDUseCase_Execute(t *testing.T) {
	t.Run("when category exists then returns category", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		expectedCategory := &models.Category{
			ID:          categoryID,
			Name:        "Electronics",
			Description: "Electronic products",
			Image: &models.Image{
				ID:  1,
				URL: "https://cloudinary.com/image.jpg",
			},
		}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(expectedCategory, nil)

		useCase := NewGetByIDUseCase(categoryServiceMock)

		// Act
		result, err := useCase.Execute(ctx, categoryID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedCategory.ID, result.ID)
		assert.Equal(t, expectedCategory.Name, result.Name)
		assert.NotNil(t, result.Image)
	})

	t.Run("when category not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 999
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.CategoryNotFound}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(nil, notFoundError)

		useCase := NewGetByIDUseCase(categoryServiceMock)

		// Act
		result, err := useCase.Execute(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.CategoryNotFound, notFound.Message)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		expectedError := errors.New("database error")

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(nil, expectedError)

		useCase := NewGetByIDUseCase(categoryServiceMock)

		// Act
		result, err := useCase.Execute(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when category has no image then returns category without image", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		expectedCategory := &models.Category{
			ID:          categoryID,
			Name:        "Electronics",
			Description: "Electronic products",
			Image:       nil,
		}

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(expectedCategory, nil)

		useCase := NewGetByIDUseCase(categoryServiceMock)

		// Act
		result, err := useCase.Execute(ctx, categoryID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Image)
	})
}
