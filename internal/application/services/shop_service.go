package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// Image type constants
const (
	imageTypeLogo  = "logo"
	imageTypeCover = "cover"
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

// GetBySlug retrieves a shop by slug.
// Used for public store endpoints.
func (s *ShopService) GetBySlug(ctx context.Context, slug string) (*models.Shop, error) {
	return s.shopRepository.GetBySlug(ctx, slug)
}

// GetActivePaymentMethod retrieves an active payment method from the shop.
// Returns ValidationError if not found or inactive.
func (s *ShopService) GetActivePaymentMethod(shop *models.Shop, paymentMethodID int) (*models.PaymentMethod, error) {
	for _, method := range shop.PaymentMethods {
		if method.ID == paymentMethodID {
			if !method.IsActive {
				return nil, &errors.ValidationError{Message: errors.PaymentMethodNotFound}
			}
			return method, nil
		}
	}
	return nil, &errors.ValidationError{Message: errors.PaymentMethodNotFound}
}

// GetActiveDeliveryMethod retrieves an active delivery method from the shop.
// Returns ValidationError if not found or inactive.
func (s *ShopService) GetActiveDeliveryMethod(shop *models.Shop, deliveryMethodID int) (*models.DeliveryMethod, error) {
	for _, method := range shop.DeliveryMethods {
		if method.ID == deliveryMethodID {
			if !method.IsActive {
				return nil, &errors.ValidationError{Message: errors.DeliveryMethodNotFound}
			}
			return method, nil
		}
	}
	return nil, &errors.ValidationError{Message: errors.DeliveryMethodNotFound}
}

// ValidateShippingCost validates the shipping cost matches the delivery method configuration.
// Returns ValidationError if mismatch.
//
//nolint:gocyclo // Complexity is intentional - sequential validation for each delivery type (pickup, fixed, zones, free)
func (s *ShopService) ValidateShippingCost(shippingCost float64, method *models.DeliveryMethod, selectedZoneID int) error {
	// Pickup should have zero shipping cost
	if method.Code == models.DeliveryMethodPickup {
		if shippingCost != 0 {
			return &errors.ValidationError{Message: errors.PickupShouldHaveZeroCost}
		}
		return nil
	}

	// Delivery method - check for fixed price first
	if method.DeliveryConfig != nil && method.DeliveryConfig.FixedPrice != nil {
		if shippingCost != *method.DeliveryConfig.FixedPrice {
			return &errors.ValidationError{Message: errors.ShippingCostMismatch}
		}
		return nil
	}

	// Zone-based pricing
	if len(method.DeliveryZones) > 0 {
		if selectedZoneID == 0 {
			return &errors.ValidationError{Message: errors.DeliveryZoneRequired}
		}

		zone := s.findDeliveryZone(method.DeliveryZones, selectedZoneID)
		if zone == nil {
			return &errors.ValidationError{Message: errors.DeliveryZoneNotFound}
		}

		if shippingCost != zone.Price {
			return &errors.ValidationError{Message: errors.ShippingCostMismatch}
		}
		return nil
	}

	// No configuration - shipping cost should be zero (free delivery)
	if shippingCost != 0 {
		return &errors.ValidationError{Message: errors.ShippingCostMismatch}
	}

	return nil
}

// findDeliveryZone searches for a delivery zone by ID in the method's zones.
func (s *ShopService) findDeliveryZone(zones []*models.DeliveryZone, id int) *models.DeliveryZone {
	for _, zone := range zones {
		if zone.ID == id {
			return zone
		}
	}
	return nil
}

// validateShop validates business rules for shop update.
// Validates nested models like PaymentMethods with TransferConfig.
func (s *ShopService) validateShop(shop *models.Shop) error {
	// Validate shop domain model (e.g., primary_color format)
	if err := shop.Validate(); err != nil {
		return err
	}

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
			img.Type = imageTypeLogo
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
			img.Type = imageTypeCover
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

	// 2. Filter out existing images of the same type as new uploads
	// This ensures the old image will be:
	// - Removed from the JSONB sent to the stored procedure
	// - Detected as "to delete" by the SP (exists in DB but not in JSONB)
	// - Its storage_ref returned for cleanup from Cloudinary
	if logoImage != nil || coverImage != nil {
		filteredImages := make([]*models.Image, 0, len(shop.Images))
		for _, img := range shop.Images {
			keepLogo := img.Type != imageTypeLogo || logoImage == nil
			keepCover := img.Type != imageTypeCover || coverImage == nil
			if keepLogo && keepCover {
				filteredImages = append(filteredImages, img)
			}
		}
		shop.Images = filteredImages
	}

	// 3. Append uploaded images to shop (order: logo first, then cover)
	if logoImage != nil {
		shop.Images = append(shop.Images, logoImage)
	}
	if coverImage != nil {
		shop.Images = append(shop.Images, coverImage)
	}

	// 4. Persist to database - returns refs of deleted images
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

	// 5. Cleanup: delete removed images from storage (fire-and-forget)
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
