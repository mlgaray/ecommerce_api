package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type UpdateShopUseCase interface {
	// Execute updates a shop with optional new images.
	// newLogoBuffer and newCoverBuffer are optional (nil if no new image).
	Execute(ctx context.Context, shopID int, shop *models.Shop, newLogoBuffer []byte, newCoverBuffer []byte) error
}
