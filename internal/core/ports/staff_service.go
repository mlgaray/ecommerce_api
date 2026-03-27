package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type StaffService interface {
	// Auth: used by sign-in use case
	GetShopRolesByUserID(ctx context.Context, userID int) ([]claims.ShopRole, error)

	// CRUD: used by staff handler
	Create(ctx context.Context, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error)
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) ([]*models.Staff, error)
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) (int, error)
	GetByID(ctx context.Context, staffID, shopID int) (*models.Staff, error)
	Update(ctx context.Context, staffID, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error)
	Delete(ctx context.Context, staffID, shopID int) error
	ToggleStatus(ctx context.Context, staffID, shopID int) (*models.Staff, error)
}
