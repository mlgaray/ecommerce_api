package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type RoleRepository interface {
	GetByID(ctx context.Context, id int) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
}
