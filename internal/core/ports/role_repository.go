package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*models.Role, error)
}
