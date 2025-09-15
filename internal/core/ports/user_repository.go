package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	AssignRole(ctx context.Context, userID int, roleID int) error
}
