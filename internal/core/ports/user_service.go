package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type UserService interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ValidateCredentials(ctx context.Context, user *models.User, password string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
}
