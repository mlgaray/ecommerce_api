package stubs

import (
	"context"
	"fmt"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// AssetService is a stub implementation for integration tests.
// It simulates upload/delete operations without external dependencies.
type AssetService struct {
	uploadCounter int
}

// NewAssetService creates a new stub AssetService.
func NewAssetService() *AssetService {
	return &AssetService{}
}

// Upload simulates uploading an image and returns a stub result.
func (s *AssetService) Upload(ctx context.Context, buffer []byte, folder string) (*models.Image, error) {
	s.uploadCounter++
	return &models.Image{
		URL:        fmt.Sprintf("https://stub-storage.com/%s/image_%d.jpg", folder, s.uploadCounter),
		StorageRef: fmt.Sprintf("%s/stub_ref_%d", folder, s.uploadCounter),
	}, nil
}

// UploadMultiple simulates uploading multiple images and returns stub results.
func (s *AssetService) UploadMultiple(ctx context.Context, buffers [][]byte, folder string) ([]*models.Image, error) {
	results := make([]*models.Image, len(buffers))
	for i := range buffers {
		s.uploadCounter++
		results[i] = &models.Image{
			URL:        fmt.Sprintf("https://stub-storage.com/%s/image_%d.jpg", folder, s.uploadCounter),
			StorageRef: fmt.Sprintf("%s/stub_ref_%d", folder, s.uploadCounter),
		}
	}
	return results, nil
}

// Delete simulates deleting an image (no-op for tests).
func (s *AssetService) Delete(ctx context.Context, storageRef string) error {
	return nil
}
