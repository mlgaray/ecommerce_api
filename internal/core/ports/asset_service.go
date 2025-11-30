package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// AssetService defines the interface for asset storage operations.
// Returns domain models - infrastructure adapts to domain needs.
type AssetService interface {
	// Upload uploads an image to the storage provider.
	// folder: path like "shop_1/categories" or "shop_1/products"
	Upload(ctx context.Context, buffer []byte, folder string) (*models.Image, error)

	// UploadMultiple uploads multiple images in parallel to the storage provider.
	// On error, rolls back successfully uploaded images automatically.
	UploadMultiple(ctx context.Context, buffers [][]byte, folder string) ([]*models.Image, error)

	// Delete removes an image from the storage provider.
	// storageRef: the provider-specific reference returned by Upload
	Delete(ctx context.Context, storageRef string) error
}
