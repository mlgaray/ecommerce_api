package product

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetByIDUseCase struct {
	productService ports.ProductService
}

func NewGetByIDUseCase(productService ports.ProductService) ports.GetByIDUseCase {
	return &GetByIDUseCase{
		productService: productService,
	}
}

func (uc *GetByIDUseCase) Execute(ctx context.Context, productID int) (*models.Product, error) {
	return uc.productService.GetByID(ctx, productID)
}
