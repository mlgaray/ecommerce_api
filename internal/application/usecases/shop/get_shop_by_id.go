package shop

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetShopByIDUseCase struct {
	shopRepository ports.ShopRepository
}

func NewGetShopByIDUseCase(shopRepository ports.ShopRepository) ports.GetShopByIDUseCase {
	return &GetShopByIDUseCase{
		shopRepository: shopRepository,
	}
}

func (uc *GetShopByIDUseCase) Execute(ctx context.Context, shopID int) (*models.Shop, error) {
	return uc.shopRepository.GetByID(ctx, shopID)
}
