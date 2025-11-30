package assets

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Log field constants
const (
	CloudinaryAssetServiceField   = "cloudinary_asset_service"
	UploadFunctionField           = "upload"
	UploadMultipleFunctionField   = "upload_multiple"
	DeleteFunctionField           = "delete"
	CloudinaryUploadSubFuncField  = "cloudinary.Upload.Upload"
	CloudinaryDestroySubFuncField = "cloudinary.Upload.Destroy"
)

// Cloudinary defines the interface for Cloudinary upload operations.
// This interface is implicitly implemented by *uploader.API from Cloudinary SDK.
type Cloudinary interface {
	Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error)
	Destroy(ctx context.Context, params uploader.DestroyParams) (*uploader.DestroyResult, error)
}

// CloudinaryAssetService implements AssetService using Cloudinary.
type CloudinaryAssetService struct {
	cloudinary Cloudinary
}

// NewCloudinaryAssetService creates a new Cloudinary asset service.
func NewCloudinaryAssetService(conn CloudinaryConnection) *CloudinaryAssetService {
	return &CloudinaryAssetService{
		cloudinary: &conn.Connect().Upload,
	}
}

// Upload uploads an image to Cloudinary.
func (s *CloudinaryAssetService) Upload(ctx context.Context, buffer []byte, folder string) (*models.Image, error) {
	result, err := s.cloudinary.Upload(ctx, bytes.NewReader(buffer), uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CloudinaryAssetServiceField,
			"function": UploadFunctionField,
			"sub_func": CloudinaryUploadSubFuncField,
			"folder":   folder,
			"error":    err.Error(),
		}).Error("Failed to upload image to Cloudinary")
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	return &models.Image{
		URL:        result.SecureURL,
		StorageRef: result.PublicID,
	}, nil
}

// UploadMultiple uploads multiple images in parallel to Cloudinary.
// On error, rolls back successfully uploaded images automatically.
func (s *CloudinaryAssetService) UploadMultiple(ctx context.Context, buffers [][]byte, folder string) ([]*models.Image, error) {
	if len(buffers) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	results := make([]*models.Image, len(buffers))
	errors := make([]error, len(buffers))

	// Upload all images in parallel
	for i, buffer := range buffers {
		wg.Add(1)
		go func(idx int, buf []byte) {
			defer wg.Done()
			result, err := s.Upload(ctx, buf, folder)
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = result
		}(i, buffer)
	}

	wg.Wait()

	// Check for errors and collect successful uploads
	var firstError error
	successfulRefs := make([]string, 0, len(buffers))
	for i, err := range errors {
		if err != nil {
			if firstError == nil {
				firstError = err
			}
		} else if results[i] != nil {
			successfulRefs = append(successfulRefs, results[i].StorageRef)
		}
	}

	// If any upload failed, rollback successful ones
	if firstError != nil {
		logs.WithFields(map[string]interface{}{
			"file":           CloudinaryAssetServiceField,
			"function":       UploadMultipleFunctionField,
			"folder":         folder,
			"total_images":   len(buffers),
			"failed_count":   countErrors(errors),
			"rollback_count": len(successfulRefs),
			"error":          firstError.Error(),
		}).Error("Failed to upload multiple images")

		for _, ref := range successfulRefs {
			_ = s.Delete(ctx, ref) // Fire-and-forget rollback
		}
		return nil, firstError
	}

	return results, nil
}

// countErrors counts non-nil errors in slice
func countErrors(errors []error) int {
	count := 0
	for _, err := range errors {
		if err != nil {
			count++
		}
	}
	return count
}

// Delete removes an image from Cloudinary.
func (s *CloudinaryAssetService) Delete(ctx context.Context, storageRef string) error {
	if storageRef == "" {
		return nil // Nothing to delete
	}

	_, err := s.cloudinary.Destroy(ctx, uploader.DestroyParams{
		PublicID: storageRef,
	})
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CloudinaryAssetServiceField,
			"function":    DeleteFunctionField,
			"sub_func":    CloudinaryDestroySubFuncField,
			"storage_ref": storageRef,
			"error":       err.Error(),
		}).Error("Failed to delete image from Cloudinary")
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}
