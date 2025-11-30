package assets

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestMain(m *testing.M) {
	logs.Init()
	code := m.Run()
	os.Exit(code)
}

func TestCloudinaryAssetService_Upload(t *testing.T) {
	t.Run("when upload succeeds then returns UploadResult", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		buffer := []byte("fake image data")
		folder := "products"

		mockCloudinary.EXPECT().
			Upload(ctx, mock.Anything, uploader.UploadParams{Folder: folder}).
			Return(&uploader.UploadResult{
				SecureURL: "https://cloudinary.com/products/image.jpg",
				PublicID:  "products/image123",
			}, nil)

		// Act
		result, err := service.Upload(ctx, buffer, folder)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "https://cloudinary.com/products/image.jpg", result.URL)
		assert.Equal(t, "products/image123", result.StorageRef)
	})

	t.Run("when upload fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		buffer := []byte("fake image data")
		folder := "products"

		mockCloudinary.EXPECT().
			Upload(ctx, mock.Anything, uploader.UploadParams{Folder: folder}).
			Return(nil, errors.New("cloudinary connection failed"))

		// Act
		result, err := service.Upload(ctx, buffer, folder)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to upload image")
	})
}

func TestCloudinaryAssetService_UploadMultiple(t *testing.T) {
	t.Run("when all uploads succeed then returns all results", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		buffers := [][]byte{
			[]byte("image1 data"),
			[]byte("image2 data"),
		}
		folder := "products"

		// Since uploads run in parallel goroutines, order is non-deterministic
		// Use same return value for both calls
		mockCloudinary.EXPECT().
			Upload(ctx, mock.Anything, uploader.UploadParams{Folder: folder}).
			Return(&uploader.UploadResult{
				SecureURL: "https://cloudinary.com/products/image.jpg",
				PublicID:  "products/image",
			}, nil).Times(2)

		// Act
		results, err := service.UploadMultiple(ctx, buffers, folder)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.NotNil(t, results[0])
		assert.NotNil(t, results[1])
	})

	t.Run("when empty buffers then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		// Act
		results, err := service.UploadMultiple(ctx, [][]byte{}, "products")

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, results)
	})

	t.Run("when one upload fails then rolls back successful uploads and returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		buffers := [][]byte{
			[]byte("image1 data"),
			[]byte("image2 data"),
		}
		folder := "products"

		// First upload succeeds
		mockCloudinary.EXPECT().
			Upload(ctx, mock.Anything, uploader.UploadParams{Folder: folder}).
			Return(&uploader.UploadResult{
				SecureURL: "https://cloudinary.com/products/image1.jpg",
				PublicID:  "products/image1",
			}, nil).Once()

		// Second upload fails
		mockCloudinary.EXPECT().
			Upload(ctx, mock.Anything, uploader.UploadParams{Folder: folder}).
			Return(nil, errors.New("upload failed")).Once()

		// Expect rollback of successful upload
		mockCloudinary.EXPECT().
			Destroy(ctx, uploader.DestroyParams{PublicID: "products/image1"}).
			Return(&uploader.DestroyResult{}, nil).Once()

		// Act
		results, err := service.UploadMultiple(ctx, buffers, folder)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to upload image")
	})
}

func TestCloudinaryAssetService_Delete(t *testing.T) {
	t.Run("when delete succeeds then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		storageRef := "products/image123"

		mockCloudinary.EXPECT().
			Destroy(ctx, uploader.DestroyParams{PublicID: storageRef}).
			Return(&uploader.DestroyResult{}, nil)

		// Act
		err := service.Delete(ctx, storageRef)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when storageRef is empty then returns nil without calling destroy", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		// Act
		err := service.Delete(ctx, "")

		// Assert
		assert.NoError(t, err)
		// No expectations set, so mock will fail if Destroy is called
	})

	t.Run("when delete fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockCloudinary := mocks.NewCloudinary(t)
		service := &CloudinaryAssetService{cloudinary: mockCloudinary}

		storageRef := "products/image123"

		mockCloudinary.EXPECT().
			Destroy(ctx, uploader.DestroyParams{PublicID: storageRef}).
			Return(nil, errors.New("cloudinary delete failed"))

		// Act
		err := service.Delete(ctx, storageRef)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete image")
	})
}

func TestCountErrors(t *testing.T) {
	t.Run("when no errors then returns 0", func(t *testing.T) {
		errs := []error{nil, nil, nil}
		assert.Equal(t, 0, countErrors(errs))
	})

	t.Run("when some errors then returns count", func(t *testing.T) {
		errs := []error{
			errors.New("error1"),
			nil,
			errors.New("error2"),
		}
		assert.Equal(t, 2, countErrors(errs))
	})

	t.Run("when all errors then returns total count", func(t *testing.T) {
		errs := []error{
			errors.New("error1"),
			errors.New("error2"),
		}
		assert.Equal(t, 2, countErrors(errs))
	})
}
