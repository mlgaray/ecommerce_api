package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ShopRepository interface {
	Create(ctx context.Context, shop *models.Shop) (*models.Shop, error)
	// GetShopsByUserID returns all shops owned by a user.
	// Used during authentication to include shop IDs in JWT token.
	GetShopsByUserID(ctx context.Context, userID int) ([]*models.Shop, error)
}
