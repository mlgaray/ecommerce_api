package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type SignInUseCase interface {
	Execute(ctx context.Context, user *models.User) (string, error)
}
