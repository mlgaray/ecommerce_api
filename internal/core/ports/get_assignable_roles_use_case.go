package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type GetAssignableRolesUseCase interface {
	Execute(ctx context.Context) ([]*models.Role, error)
}
