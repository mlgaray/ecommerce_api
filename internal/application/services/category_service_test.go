package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// Note: Category validation (name, description) is handled at the HTTP layer (contracts).
// The service layer focuses on orchestration: upload, persist, rollback.

// =============================================================================
// Test Helpers
// =============================================================================

// newValidCategory creates a valid category for testing
func newValidCategory() *models.Category {
	return &models.Category{
		Name:        "Test Category",
		Description: "Test Description",
	}
}

// newValidCategoryWithID creates a valid category with ID 1 for testing
func newValidCategoryWithID() *models.Category {
	c := newValidCategory()
	c.ID = 1
	return c
}

// =============================================================================
// Create Tests
// =============================================================================

func TestCategoryService_Create(t *testing.T) {
	t.Run("when category is valid with image then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		imageBuffer := []byte("image_data")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/image.jpg",
			StorageRef: "shop_1/categories/abc123",
		}

		expectedCategory := newValidCategoryWithID()
		expectedCategory.Image = &models.Image{
			URL:        uploadResult.URL,
			StorageRef: uploadResult.StorageRef,
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_1/categories").
			Return(uploadResult, nil)

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Category"), shopID).
			Return(expectedCategory, nil)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedCategory.ID, result.ID)
		assert.NotNil(t, result.Image)
		assert.Equal(t, uploadResult.URL, result.Image.URL)
	})

	t.Run("when category is valid without image then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		var imageBuffer []byte // No image

		expectedCategory := newValidCategoryWithID()

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, category, shopID).
			Return(expectedCategory, nil)

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called when no image

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedCategory.ID, result.ID)
	})

	t.Run("when asset upload fails then returns error without calling repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		imageBuffer := []byte("image_data")
		uploadError := stdErrors.New("upload failed")

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_1/categories").
			Return(nil, uploadError)

		repoMock := mocks.NewCategoryRepository(t)
		// Repository should NOT be called when upload fails

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, uploadError, err)
	})

	t.Run("when repository fails then deletes uploaded image and returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		imageBuffer := []byte("image_data")
		repoError := stdErrors.New("database error")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/image.jpg",
			StorageRef: "shop_1/categories/abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, uploadResult.StorageRef).
			Return(nil) // Rollback succeeds

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Category"), shopID).
			Return(nil, repoError)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)
	})

	t.Run("when repository fails and image delete fails then still returns repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		imageBuffer := []byte("image_data")
		repoError := stdErrors.New("database error")
		deleteError := stdErrors.New("delete failed")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/image.jpg",
			StorageRef: "shop_1/categories/abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, uploadResult.StorageRef).
			Return(deleteError) // Rollback fails but we ignore it

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Category"), shopID).
			Return(nil, repoError)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err) // Original error, not delete error
	})

	t.Run("when category already exists then returns duplicate error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		imageBuffer := []byte("image_data")
		duplicateError := &errors.DuplicateRecordError{Message: errors.CategoryAlreadyExistsInShop}

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/image.jpg",
			StorageRef: "shop_1/categories/abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, uploadResult.StorageRef).
			Return(nil)

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Category"), shopID).
			Return(nil, duplicateError)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var dupErr *errors.DuplicateRecordError
		assert.True(t, stdErrors.As(err, &dupErr))
		assert.Equal(t, errors.CategoryAlreadyExistsInShop, dupErr.Message)
	})

	t.Run("when different shop_ids then uses correct folder path", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 42
		category := newValidCategory()
		imageBuffer := []byte("image_data")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/image.jpg",
			StorageRef: "shop_42/categories/abc123",
		}

		expectedCategory := newValidCategoryWithID()

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, imageBuffer, "shop_42/categories"). // Correct folder
			Return(uploadResult, nil)

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Category"), shopID).
			Return(expectedCategory, nil)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCategoryService_EdgeCases(t *testing.T) {
	t.Run("when empty image buffer then does not upload", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		emptyBuffer := []byte{} // Empty but not nil

		expectedCategory := newValidCategoryWithID()

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, category, shopID).
			Return(expectedCategory, nil)

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called for empty buffer

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, emptyBuffer, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when repository fails without image then no rollback needed", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		category := newValidCategory()
		var imageBuffer []byte // No image
		repoError := stdErrors.New("database error")

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Create(ctx, category, shopID).
			Return(nil, repoError)

		assetMock := mocks.NewAssetService(t)
		// Delete should NOT be called when there's no image to rollback

		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, category, imageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)
	})
}
