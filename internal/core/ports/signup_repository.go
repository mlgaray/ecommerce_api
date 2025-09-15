package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type SignupRepository interface {
	CreateUserWithShop(ctx context.Context, user *models.User, shop *models.Shop) (*models.User, error)
}
