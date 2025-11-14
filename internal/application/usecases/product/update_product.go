package product

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type UpdateProductUseCase struct {
	productService ports.ProductService
}

func NewUpdateProductUseCase(productService ports.ProductService) ports.UpdateProductUseCase {
	return &UpdateProductUseCase{
		productService: productService,
	}
}

func (uc *UpdateProductUseCase) Execute(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error {
	// Uses stored procedure for optimal performance (single DB round trip)
	return uc.productService.Update(ctx, productID, product, newImageBuffers)
}
