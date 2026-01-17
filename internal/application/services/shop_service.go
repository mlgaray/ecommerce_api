package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// ShopService contains business logic, validations, and data access coordination.
// Use Cases orchestrate the flow using this service (NOT repositories directly).
type ShopService struct {
	shopRepository ports.ShopRepository
	assetService   ports.AssetService
}

func NewShopService(shopRepository ports.ShopRepository, assetService ports.AssetService) *ShopService {
	return &ShopService{
		shopRepository: shopRepository,
		assetService:   assetService,
	}
}

// GetByID retrieves a shop by ID.
// Delegates to repository - service layer can add business logic if needed.
func (s *ShopService) GetByID(ctx context.Context, shopID int) (*models.Shop, error) {
	return s.shopRepository.GetByID(ctx, shopID)
}

// validateShop validates business rules for shop update.
// Validates nested models like PaymentMethods with TransferConfig.
func (s *ShopService) validateShop(shop *models.Shop) error {
	// Validate payment methods with transfer config
	for _, pm := range shop.PaymentMethods {
		if pm.IsActive && pm.TransferConfig != nil {
			if err := pm.TransferConfig.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Update updates an existing shop with optional new images.
// Business logic: validates shop, uploads new images to storage, persists via repository.
// Handles: upload new images -> persist to DB -> delete removed images from storage.
// Uses parallel uploads for logo and cover images via WaitGroup.
//
//nolint:gocyclo // Complexity is intentional for readability - parallel upload logic with rollback
func (s *ShopService) Update(ctx context.Context, shopID int, shop *models.Shop, newLogoBuffer []byte, newCoverBuffer []byte) error {
	// 0. Validate shop business rules
	if err := s.validateShop(shop); err != nil {
		return err
	}

	// 1. Upload new images in parallel using WaitGroup
	var (
		wg         sync.WaitGroup
		logoImage  *models.Image
		coverImage *models.Image
		logoErr    error
		coverErr   error
	)

	folder := fmt.Sprintf("shop_%d/images", shopID)

	// Upload logo in parallel
	if len(newLogoBuffer) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			img, err := s.assetService.Upload(ctx, newLogoBuffer, folder)
			if err != nil {
				logoErr = err
				return
			}
			img.Type = "logo"
			logoImage = img
		}()
	}

	// Upload cover in parallel
	if len(newCoverBuffer) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			img, err := s.assetService.Upload(ctx, newCoverBuffer, folder)
			if err != nil {
				coverErr = err
				return
			}
			img.Type = "cover"
			coverImage = img
		}()
	}

	// Wait for all uploads to complete
	wg.Wait()

	// Check for errors and rollback if needed
	if logoErr != nil || coverErr != nil {
		// Rollback: delete any successfully uploaded images
		if logoImage != nil {
			_ = s.assetService.Delete(ctx, logoImage.StorageRef)
		}
		if coverImage != nil {
			_ = s.assetService.Delete(ctx, coverImage.StorageRef)
		}
		// Return first error
		if logoErr != nil {
			return logoErr
		}
		return coverErr
	}

	// 2. Append uploaded images to shop (order: logo first, then cover)
	if logoImage != nil {
		shop.Images = append(shop.Images, logoImage)
	}
	if coverImage != nil {
		shop.Images = append(shop.Images, coverImage)
	}

	// 3. Persist to database - returns refs of deleted images
	deletedRefs, err := s.shopRepository.Update(ctx, shopID, shop)
	if err != nil {
		// Rollback: delete newly uploaded images
		if logoImage != nil {
			_ = s.assetService.Delete(ctx, logoImage.StorageRef)
		}
		if coverImage != nil {
			_ = s.assetService.Delete(ctx, coverImage.StorageRef)
		}
		return err
	}

	// 4. Cleanup: delete removed images from storage (fire-and-forget)
	for _, ref := range deletedRefs {
		if ref != "" {
			go func(storageRef string) {
				bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				_ = s.assetService.Delete(bgCtx, storageRef)
			}(ref)
		}
	}

	return nil
}
