package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

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

func (s *ProductService) Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
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
	placeholderImages := make([]string, len(imageBuffers))
	for i := range imageBuffers {
		placeholderImages[i] = "https://placeholder.com/image_" + string(rune(i+1))
	}
	product.Images = placeholderImages

	// Create product with shop association
	return s.productRepository.Create(ctx, product, shopID)
}
