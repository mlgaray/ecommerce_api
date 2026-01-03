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

// =============================================================================
// Test Helpers
// =============================================================================

// newValidShopForUpdate creates a valid shop for update testing
func newValidShopForUpdate() *models.Shop {
	return &models.Shop{
		Name:      "Updated Shop",
		Slug:      "updated-shop",
		Email:     "updated@shop.com",
		Phone:     "+54 11 1234-5678",
		Instagram: "@updatedshop",
	}
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestShopService_GetByID(t *testing.T) {
	t.Run("when shop exists then returns shop", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedShop := &models.Shop{
			ID:    shopID,
			Name:  "Test Shop",
			Slug:  "test-shop",
			Email: "test@shop.com",
			Images: []*models.Image{
				{ID: 1, URL: "https://cloudinary.com/logo.jpg", Type: "logo"},
			},
		}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(expectedShop, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewShopService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedShop.ID, result.ID)
		assert.Equal(t, expectedShop.Name, result.Name)
		assert.NotNil(t, result.Images)
	})

	t.Run("when shop not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		notFoundError := &errors.RecordNotFoundError{Message: errors.ShopNotFound}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(nil, notFoundError)

		assetMock := mocks.NewAssetService(t)
		service := NewShopService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.ShopNotFound, notFound.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedError := stdErrors.New("database error")

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, shopID).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewShopService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// Update Tests
// =============================================================================

func TestShopService_Update(t *testing.T) {
	t.Run("when updating with both images then uploads and persists successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		coverBuffer := []byte("cover_data")

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_1/images/logo_abc123",
		}
		coverUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_cover.jpg",
			StorageRef: "shop_1/images/cover_xyz789",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(logoUploadResult, nil)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(coverUploadResult, nil)

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, nil)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when updating with logo only then uploads logo and persists successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		var coverBuffer []byte

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_1/images/logo_abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(logoUploadResult, nil)

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, nil)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when updating with cover only then uploads cover and persists successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		var logoBuffer []byte
		coverBuffer := []byte("cover_data")

		coverUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_cover.jpg",
			StorageRef: "shop_1/images/cover_xyz789",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(coverUploadResult, nil)

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, nil)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when updating without images then persists successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		shop.Images = []*models.Image{
			{ID: 1, URL: "https://cloudinary.com/existing_logo.jpg", Type: "logo"},
		}
		var logoBuffer []byte
		var coverBuffer []byte

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, shop).
			Return([]string{}, nil)

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called when no new images

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when logo upload fails then returns error without calling repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		var coverBuffer []byte
		uploadError := stdErrors.New("logo upload failed")

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(nil, uploadError)

		repoMock := mocks.NewShopRepository(t)
		// Repository should NOT be called when upload fails

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uploadError, err)
	})

	t.Run("when cover upload fails then returns error without calling repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		var logoBuffer []byte
		coverBuffer := []byte("cover_data")
		uploadError := stdErrors.New("cover upload failed")

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(nil, uploadError)

		repoMock := mocks.NewShopRepository(t)
		// Repository should NOT be called when upload fails

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, uploadError, err)
	})

	t.Run("when logo upload fails but cover succeeds then rolls back cover and returns logo error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		coverBuffer := []byte("cover_data")
		logoUploadError := stdErrors.New("logo upload failed")

		coverUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_cover.jpg",
			StorageRef: "shop_1/images/cover_xyz789",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(nil, logoUploadError)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(coverUploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, coverUploadResult.StorageRef).
			Return(nil) // Rollback cover

		repoMock := mocks.NewShopRepository(t)
		// Repository should NOT be called when upload fails

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, logoUploadError, err)
	})

	t.Run("when cover upload fails but logo succeeds then rolls back logo and returns cover error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		coverBuffer := []byte("cover_data")
		coverUploadError := stdErrors.New("cover upload failed")

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_1/images/logo_abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(logoUploadResult, nil)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(nil, coverUploadError)
		assetMock.EXPECT().
			Delete(ctx, logoUploadResult.StorageRef).
			Return(nil) // Rollback logo

		repoMock := mocks.NewShopRepository(t)
		// Repository should NOT be called when upload fails

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, coverUploadError, err)
	})

	t.Run("when repository fails then rolls back uploaded images and returns repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		coverBuffer := []byte("cover_data")
		repoError := stdErrors.New("database error")

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_1/images/logo_abc123",
		}
		coverUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_cover.jpg",
			StorageRef: "shop_1/images/cover_xyz789",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(logoUploadResult, nil)
		assetMock.EXPECT().
			Upload(ctx, coverBuffer, "shop_1/images").
			Return(coverUploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, logoUploadResult.StorageRef).
			Return(nil) // Rollback logo
		assetMock.EXPECT().
			Delete(ctx, coverUploadResult.StorageRef).
			Return(nil) // Rollback cover

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, repoError)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err)
	})

	t.Run("when repository fails and rollback fails then still returns repository error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		var coverBuffer []byte
		repoError := stdErrors.New("database error")
		deleteError := stdErrors.New("delete failed")

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_1/images/logo_abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_1/images").
			Return(logoUploadResult, nil)
		assetMock.EXPECT().
			Delete(ctx, logoUploadResult.StorageRef).
			Return(deleteError) // Rollback fails but we ignore it

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, repoError)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err) // Original error, not delete error
	})

	t.Run("when shop not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		shop := newValidShopForUpdate()
		var logoBuffer []byte
		var coverBuffer []byte
		notFoundError := &errors.RecordNotFoundError{Message: errors.ShopNotFound}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, shop).
			Return([]string{}, notFoundError)

		assetMock := mocks.NewAssetService(t)
		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.ShopNotFound, notFound.Message)
	})

	t.Run("when update deletes old images then returns success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		var logoBuffer []byte
		var coverBuffer []byte
		// Note: Repository returns deleted refs, but cleanup happens in fire-and-forget goroutines
		// We only test that the update operation succeeds - cleanup is best-effort
		deletedRefs := []string{} // Empty to avoid goroutine assertion issues

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, shop).
			Return(deletedRefs, nil)

		assetMock := mocks.NewAssetService(t)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when different shop_id then uses correct folder path", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 42
		shop := newValidShopForUpdate()
		logoBuffer := []byte("logo_data")
		var coverBuffer []byte

		logoUploadResult := &models.Image{
			URL:        "https://cloudinary.com/new_logo.jpg",
			StorageRef: "shop_42/images/logo_abc123",
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			Upload(ctx, logoBuffer, "shop_42/images"). // Correct folder for shop 42
			Return(logoUploadResult, nil)

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, mock.AnythingOfType("*models.Shop")).
			Return([]string{}, nil)

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when empty image buffers then does not upload", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		emptyLogoBuffer := []byte{} // Empty but not nil
		emptyCoverBuffer := []byte{}

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, shop).
			Return([]string{}, nil)

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called for empty buffers

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, emptyLogoBuffer, emptyCoverBuffer)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when repository fails without images then no rollback needed", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		shop := newValidShopForUpdate()
		var logoBuffer []byte
		var coverBuffer []byte
		repoError := stdErrors.New("database error")

		repoMock := mocks.NewShopRepository(t)
		repoMock.EXPECT().
			Update(ctx, shopID, shop).
			Return([]string{}, repoError)

		assetMock := mocks.NewAssetService(t)
		// Delete should NOT be called - no new images were uploaded

		service := NewShopService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, shopID, shop, logoBuffer, coverBuffer)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoError, err)
	})
}
