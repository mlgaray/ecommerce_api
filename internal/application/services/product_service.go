package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// ProductService contains business logic, validations, and data access coordination
// Use Cases orchestrate the flow using this service (NOT repositories directly)
type ProductService struct {
	productRepository ports.ProductRepository
	// TODO: Add AssetService injection when ready
	// assetService ports.AssetService
}

func NewProductService(productRepository ports.ProductRepository) *ProductService {
	return &ProductService{
		productRepository: productRepository,
	}
}

// validateProduct validates business rules for product creation/update
func (s *ProductService) validateProduct(product *models.Product) error {
	return product.Validate()
}

// validateFilters validates business rules for product filters
func (s *ProductService) validateFilters(filters *models.ProductFilters) error {
	return filters.Validate()
}

// prepareImagesForCreate processes image buffers for creation
// Currently creates placeholders - will upload to AssetService in the future
// TODO: Replace with actual AssetService integration
func (s *ProductService) prepareImagesForCreate(imageBuffers [][]byte) []models.ProductImage {
	placeholderImages := make([]models.ProductImage, len(imageBuffers))
	for i := range imageBuffers {
		placeholderImages[i] = models.ProductImage{
			URL: "https://placeholder.com/image_" + string(rune(i+1)),
			// ID is 0 (omitted) - Repository will assign it on INSERT
		}
	}
	return placeholderImages
}

// prepareImagesForUpdate processes new image buffers for update
// Currently creates placeholders - will upload to AssetService in the future
// TODO: Replace with actual AssetService integration
func (s *ProductService) prepareImagesForUpdate(newImageBuffers [][]byte) []models.ProductImage {
	newImages := make([]models.ProductImage, len(newImageBuffers))
	for i := range newImageBuffers {
		newImages[i] = models.ProductImage{
			URL: "https://placeholder.com/new_image_" + string(rune(i+1)),
			// ID is 0 (omitted) - Repository will INSERT these
		}
	}
	return newImages
}

// GetAllByShopIDWithFilters retrieves products with filters
// Validates and normalizes filters (Limit, SortBy, SortOrder) - changes propagate via pointer
// Delegates to repository for data access
func (s *ProductService) GetAllByShopIDWithFilters(ctx context.Context, filters *models.ProductFilters) ([]*models.Product, error) {
	if err := s.validateFilters(filters); err != nil {
		return nil, err
	}
	return s.productRepository.GetAllByShopIDWithFilters(ctx, *filters)
}

// CountByShopIDWithFilters returns total count of products matching filters
// Delegates to repository - service layer can add business logic if needed
func (s *ProductService) CountByShopIDWithFilters(ctx context.Context, filters models.ProductFilters) (int, error) {
	return s.productRepository.CountByShopIDWithFilters(ctx, filters)
}

// Create creates a new product
// Business logic: validates product, prepares images, delegates to repository
func (s *ProductService) Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
	// Validate business rules
	if err := s.validateProduct(product); err != nil {
		return nil, err
	}

	// Prepare images (placeholder URLs for now)
	product.Images = s.prepareImagesForCreate(imageBuffers)

	// Delegate to repository
	return s.productRepository.Create(ctx, product, shopID)
}

// GetByID retrieves a product by ID
// Delegates to repository - service layer can add business logic if needed
func (s *ProductService) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	return s.productRepository.GetByID(ctx, productID)
}

// Update updates an existing product
// Business logic: validates product, prepares new images, delegates to repository
func (s *ProductService) Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error {
	// Validate business rules
	if err := s.validateProduct(product); err != nil {
		return err
	}

	// Prepare new images (placeholder URLs for now) and append to product
	if len(newImageBuffers) > 0 {
		newImages := s.prepareImagesForUpdate(newImageBuffers)
		// Append new images to existing images in product
		product.Images = append(product.Images, newImages...)
	}

	// Delegate to repository
	return s.productRepository.Update(ctx, productID, product)
}
