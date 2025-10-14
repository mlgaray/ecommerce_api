package product

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type CreateProductUseCase struct {
	productService ports.ProductService
}

func NewCreateProductUseCase(productService ports.ProductService) ports.CreateProductUseCase {
	return &CreateProductUseCase{
		productService: productService,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
	return uc.productService.Create(ctx, product, imageBuffers, shopID)
}
