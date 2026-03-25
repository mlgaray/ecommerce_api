package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type CreateStaffUseCase interface {
	Execute(ctx context.Context, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error)
}

type GetAllStaffByShopIDUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.StaffFilters) ([]*models.Staff, string, bool, *int, error)
}

type GetStaffByIDUseCase interface {
	Execute(ctx context.Context, staffID, shopID int) (*models.Staff, error)
}

type UpdateStaffUseCase interface {
	Execute(ctx context.Context, staffID, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error)
}

type DeleteStaffUseCase interface {
	Execute(ctx context.Context, staffID, shopID int) error
}

type ToggleStaffStatusUseCase interface {
	Execute(ctx context.Context, staffID, shopID int) (*models.Staff, error)
}
