package services

import (
	"context"
	"fmt"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports" // Used for interface types in struct
)

// ProductService contains business logic, validations, and data access coordination
// Use Cases orchestrate the flow using this service (NOT repositories directly)
type ProductService struct {
	productRepository ports.ProductRepository
	assetService      ports.AssetService
}

func NewProductService(productRepository ports.ProductRepository, assetService ports.AssetService) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		assetService:      assetService,
	}
}

// validateProduct validates business rules for product creation/update
func (s *ProductService) validateProduct(product *models.Product) error {
	return product.Validate()
}

// deleteUploadedImages deletes images from storage (fire-and-forget for rollback)
func (s *ProductService) deleteUploadedImages(ctx context.Context, images []*models.Image) {
	for _, img := range images {
		if img != nil && img.StorageRef != "" {
			_ = s.assetService.Delete(ctx, img.StorageRef)
		}
	}
}

// GetAllByShopIDWithFilters retrieves products with filters.
// Assumes filters are already validated by the Use Case.
// ShopID is a context parameter (not a filter), passed separately.
// Delegates to repository for data access.
func (s *ProductService) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) ([]*models.Product, error) {
	return s.productRepository.GetAllByShopIDWithFilters(ctx, shopID, filters)
}

// CountByShopIDWithFilters returns total count of products matching filters
// ShopID is a context parameter (not a filter), passed separately
// Delegates to repository - service layer can add business logic if needed
func (s *ProductService) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) (int, error) {
	return s.productRepository.CountByShopIDWithFilters(ctx, shopID, filters)
}

// Create creates a new product with images
// Business logic: validates product, uploads images to storage, persists via repository
// Handles rollback: if DB fails, deletes uploaded images from storage
func (s *ProductService) Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
	// 1. Validate business rules
	if err := s.validateProduct(product); err != nil {
		return nil, err
	}

	// 2. Upload images to storage (parallel via AssetService)
	if len(imageBuffers) > 0 {
		folder := fmt.Sprintf("shop_%d/products", shopID)
		images, err := s.assetService.UploadMultiple(ctx, imageBuffers, folder)
		if err != nil {
			return nil, err
		}
		product.Images = images
	}

	// 3. Persist to database
	created, err := s.productRepository.Create(ctx, product, shopID)
	if err != nil {
		// Rollback: delete uploaded images from storage
		s.deleteUploadedImages(ctx, product.Images)
		return nil, err
	}

	return created, nil
}

// GetByID retrieves a product by ID
// Delegates to repository - service layer can add business logic if needed
func (s *ProductService) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	return s.productRepository.GetByID(ctx, productID)
}

// Update updates an existing product with new images
// Business logic: validates product, uploads new images to storage, persists via repository
// Handles: upload new images → persist to DB → delete removed images from storage
func (s *ProductService) Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte, shopID int) error {
	// 1. Validate business rules
	if err := s.validateProduct(product); err != nil {
		return err
	}

	// 2. Upload new images to storage (parallel via AssetService)
	if len(newImageBuffers) > 0 {
		folder := fmt.Sprintf("shop_%d/products", shopID)
		newImages, err := s.assetService.UploadMultiple(ctx, newImageBuffers, folder)
		if err != nil {
			return err
		}
		// Append new images to existing images in product
		product.Images = append(product.Images, newImages...)
	}

	// 3. Persist to database - returns refs of deleted images
	deletedRefs, err := s.productRepository.Update(ctx, productID, product, shopID)
	if err != nil {
		return err
	}

	// 4. Cleanup: delete removed images from storage (fire-and-forget)
	for _, ref := range deletedRefs {
		if ref != "" {
			_ = s.assetService.Delete(ctx, ref)
		}
	}

	return nil
}
