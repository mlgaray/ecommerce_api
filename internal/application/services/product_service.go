package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type ProductService struct {
	productRepository ports.ProductRepository
	paginationService ports.PaginationService[*models.Product]
	// TODO: Add AssetService injection when ready
	// assetService 1ports.AssetService
}

func NewProductService(productRepository ports.ProductRepository, paginationService ports.PaginationService[*models.Product]) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		paginationService: paginationService,
	}
}

func (s *ProductService) Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
	// Validate business rules (domain validation)
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// TODO: Upload images using AssetService and set URLs in product
	// For now, we'll create a placeholder for where image URLs would be stored
	//
	// Example when AssetService is ready:
	// imageURLs := make([]string, len(imageBuffers))
	// for i, buffer := range imageBuffers {
	//     uploadResult, err := s.assetService.UploadImage(ctx, buffer)
	//     if err != nil {
	//         return nil, err
	//     }
	//     imageURLs[i] = uploadResult.SecureURL
	// }
	// product.Images = imageURLs

	// For now, just set placeholder URLs
	placeholderImages := make([]models.ProductImage, len(imageBuffers))
	for i := range imageBuffers {
		placeholderImages[i] = models.ProductImage{
			URL: "https://placeholder.com/image_" + string(rune(i+1)),
			// ID is 0 (omitted) - Repository will assign it on INSERT
		}
	}
	product.Images = placeholderImages

	// Create product with shop association (uses stored procedures for optimal performance)
	return s.productRepository.Create(ctx, product, shopID)
}

func (s *ProductService) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
	// Get products from repository
	products, err := s.productRepository.GetAllByShopID(ctx, shopID, limit, cursor)
	if err != nil {
		return nil, 0, false, err
	}

	nextCursor, hasMore := s.paginationService.BuildCursorPagination(products, limit)

	return products, nextCursor, hasMore, nil
}

func (s *ProductService) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	// Get product from repository
	return s.productRepository.GetByID(ctx, productID)
}

func (s *ProductService) Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error {
	// Validate business rules (domain validation)
	if err := product.Validate(); err != nil {
		return err
	}

	// Process new images (upload when AssetService is ready)
	// TODO: When AssetService is implemented:
	// for i, buffer := range newImageBuffers {
	//     uploadResult, err := s.assetService.UploadImage(ctx, buffer)
	//     if err != nil {
	//         return err
	//     }
	//     product.Images = append(product.Images, models.ProductImage{
	//         URL: uploadResult.SecureURL,
	//     })
	// }

	// For now, create placeholders for new images
	for i := range newImageBuffers {
		product.Images = append(product.Images, models.ProductImage{
			URL: "https://placeholder.com/new_image_" + string(rune(i+1)),
			// ID is 0 (omitted) - Repository will INSERT these
		})
	}

	// Update product via repository (uses stored procedures for optimal performance)
	return s.productRepository.Update(ctx, productID, product)
}
