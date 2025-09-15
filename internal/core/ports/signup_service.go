package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type SignUpService interface {
	SignUp(ctx context.Context, user *models.User, shop *models.Shop) (*models.User, error)
}
