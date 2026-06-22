package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type RoleService interface {
	GetAllAssignable(ctx context.Context) ([]*models.Role, error)
}
