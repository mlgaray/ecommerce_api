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

// =============================================================================
// GetByID Tests
// =============================================================================

func TestCategoryService_GetByID(t *testing.T) {
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

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(expectedCategory, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, categoryID)

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
		notFoundError := &errors.RecordNotFoundError{Message: errors.CategoryNotFound}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(nil, notFoundError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, categoryID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.CategoryNotFound, notFound.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		expectedError := stdErrors.New("database error")

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, categoryID)

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
			Image:       nil, // No image
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, categoryID).
			Return(expectedCategory, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, categoryID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Image)
	})
}

// =============================================================================
// GetAllByShopIDWithFilters Tests
// =============================================================================
// Note: Validation and normalization tests have been moved to the Use Case layer.
// The service now assumes filters are already validated by the Use Case.

func TestCategoryService_GetAllByShopIDWithFilters(t *testing.T) {
	t.Run("when called then delegates to repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
		expectedCategories := []*models.Category{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Clothing"},
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(expectedCategories, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CategoryFilters{
			Limit: 20,
		}
		expectedError := stdErrors.New("database error")

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when no categories found then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		filters := models.CategoryFilters{
			Limit: 20,
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return([]*models.Category{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("when search filter applied then passes to repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		searchTerm := "Electronics"
		filters := models.CategoryFilters{
			Search: &searchTerm,
			Limit:  20,
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return([]*models.Category{{ID: 1, Name: "Electronics"}}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

// =============================================================================
// CountByShopIDWithFilters Tests
// =============================================================================

func TestCategoryService_CountByShopIDWithFilters(t *testing.T) {
	t.Run("when filters are valid then returns count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CategoryFilters{
			Limit: 20,
		}
		expectedCount := 42

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, filters).
			Return(expectedCount, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, result)
	})

	t.Run("when no categories then returns zero", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		filters := models.CategoryFilters{
			Limit: 20,
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, filters).
			Return(0, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CategoryFilters{
			Limit: 20,
		}
		expectedError := stdErrors.New("count query failed")

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, filters).
			Return(0, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when search filter applied then passes to repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		searchTerm := "Electronics"
		filters := models.CategoryFilters{
			Search: &searchTerm,
			Limit:  20,
		}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.CategoryFilters) bool {
				return f.Search != nil && *f.Search == searchTerm
			})).
			Return(5, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 5, result)
	})
}

// =============================================================================
// Update Tests
// =============================================================================

func TestCategoryService_Update(t *testing.T) {
	t.Run("when updating with new image then uploads and replaces old image", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")
		oldStorageRef := "shop_1/categories/old_ref"

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_image.jpg",
			StorageRef: "shop_1/categories/new_ref",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, oldStorageRef).
			Return(nil) // Cleanup old image

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, mock.AnythingOfType("*models.Category"), shopID).
			Return(oldStorageRef, nil)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when updating without new image then does not upload", func(t *testing.T) {
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

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, category, shopID).
			Return("", nil) // No old image to delete

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called when no new image

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when upload fails then returns error without calling repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")
		uploadError := stdErrors.New("upload failed")

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_1/categories").
			Return(nil, uploadError)

		repoMock := mocks.NewCategoryRepository(t)
		// Repository should NOT be called when upload fails

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uploadError, err)
	})

	t.Run("when repository fails then rolls back new image upload", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")
		repoError := stdErrors.New("database error")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_image.jpg",
			StorageRef: "shop_1/categories/new_ref",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, uploadResult.StorageRef).
			Return(nil) // Rollback new image

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, mock.AnythingOfType("*models.Category"), shopID).
			Return("", repoError)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err)
	})

	t.Run("when repository fails and rollback fails then returns repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")
		repoError := stdErrors.New("database error")
		deleteError := stdErrors.New("delete failed")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_image.jpg",
			StorageRef: "shop_1/categories/new_ref",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, uploadResult.StorageRef).
			Return(deleteError) // Rollback fails but we ignore it

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, mock.AnythingOfType("*models.Category"), shopID).
			Return("", repoError)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err) // Original error, not delete error
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
		notFoundError := &errors.RecordNotFoundError{Message: errors.CategoryNotFound}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, category, shopID).
			Return("", notFoundError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.CategoryNotFound, notFound.Message)
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
		duplicateError := &errors.DuplicateRecordError{Message: errors.CategoryAlreadyExistsInShop}

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, category, shopID).
			Return("", duplicateError)

		assetMock := mocks.NewAssetService(t)
		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		var dupErr *errors.DuplicateRecordError
		assert.True(t, stdErrors.As(err, &dupErr))
		assert.Equal(t, errors.CategoryAlreadyExistsInShop, dupErr.Message)
	})

	t.Run("when no old image to cleanup then does not call delete", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 1
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_image.jpg",
			StorageRef: "shop_1/categories/new_ref",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_1/categories").
			Return(uploadResult, nil)
		// Delete should NOT be called for old image cleanup (empty deletedRef)

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, mock.AnythingOfType("*models.Category"), shopID).
			Return("", nil) // Empty string = no old image to delete

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when different shop_id then uses correct folder path", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		categoryID := 1
		shopID := 42
		category := &models.Category{
			Name:        "Updated Category",
			Description: "Updated Description",
		}
		newImageBuffer := []byte("new_image_data")

		uploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_image.jpg",
			StorageRef: "shop_42/categories/new_ref",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, newImageBuffer, "shop_42/categories"). // Correct folder
			Return(uploadResult, nil)

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, mock.AnythingOfType("*models.Category"), shopID).
			Return("", nil)

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when repository fails without new image then no rollback needed", func(t *testing.T) {
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
		var newImageBuffer []byte
		repoError := stdErrors.New("database error")

		repoMock := mocks.NewCategoryRepository(t)
		repoMock.EXPECT().
			Update(ctx, categoryID, category, shopID).
			Return("", repoError)

		assetMock := mocks.NewAssetService(t)
		// Delete should NOT be called - no new image was uploaded

		service := NewCategoryService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, categoryID, category, newImageBuffer, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err)
	})
}
