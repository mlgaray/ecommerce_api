package product

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetAllByShopIDUseCase struct {
	productService ports.ProductService
}

func NewGetAllByShopIDUseCase(productService ports.ProductService) ports.GetAllByShopIDUseCase {
	return &GetAllByShopIDUseCase{
		productService: productService,
	}
}

func (uc *GetAllByShopIDUseCase) Execute(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
	return uc.productService.GetAllByShopID(ctx, shopID, limit, cursor)
}
