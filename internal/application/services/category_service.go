package services

import (
	"context"
	"fmt"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CategoryService contains business logic for category operations.
// Handles validation, image upload orchestration, and persistence.
type CategoryService struct {
	categoryRepository ports.CategoryRepository
	assetService       ports.AssetService
}

// NewCategoryService creates a new CategoryService instance.
func NewCategoryService(categoryRepository ports.CategoryRepository, assetService ports.AssetService) *CategoryService {
	return &CategoryService{
		categoryRepository: categoryRepository,
		assetService:       assetService,
	}
}

// Create creates a new category with image.
// Uploads image to storage and persists via repository.
// Handles rollback if persistence fails after image upload.
func (s *CategoryService) Create(ctx context.Context, category *models.Category, imageBuffer []byte, shopID int) (*models.Category, error) {
	// 1. Upload image to storage (if provided)
	if len(imageBuffer) > 0 {
		folder := fmt.Sprintf("shop_%d/categories", shopID)
		uploadResult, err := s.assetService.Upload(ctx, imageBuffer, folder)
		if err != nil {
			return nil, err
		}
		category.Image = &models.Image{
			URL:        uploadResult.URL,
			StorageRef: uploadResult.StorageRef,
		}
	}

	// 2. Persist category + image atomically (via SP)
	created, err := s.categoryRepository.Create(ctx, category, shopID)
	if err != nil {
		// Rollback: delete from storage if upload succeeded
		if category.Image != nil && category.Image.StorageRef != "" {
			_ = s.assetService.Delete(ctx, category.Image.StorageRef)
		}
		return nil, err
	}

	return created, nil
}
