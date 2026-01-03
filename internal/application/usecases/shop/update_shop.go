package shop

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// UpdateShopUseCase orchestrates shop update flow.
// Delegates business logic to ShopService (NOT repository directly).
type UpdateShopUseCase struct {
	shopService ports.ShopService
}

func NewUpdateShopUseCase(shopService ports.ShopService) ports.UpdateShopUseCase {
	return &UpdateShopUseCase{
		shopService: shopService,
	}
}

// Execute updates a shop with optional new images.
// Orchestration only - business logic handled by ShopService.
func (uc *UpdateShopUseCase) Execute(ctx context.Context, shopID int, shop *models.Shop, newLogoBuffer []byte, newCoverBuffer []byte) error {
	return uc.shopService.Update(ctx, shopID, shop, newLogoBuffer, newCoverBuffer)
}
