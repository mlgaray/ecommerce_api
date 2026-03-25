package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type StaffRepository interface {
	// Internal: used by signup and signin
	Create(ctx context.Context, userID, shopID int) (*models.Staff, error)
	AssignRole(ctx context.Context, staffID, roleID int) error
	GetShopRolesByUserID(ctx context.Context, userID int) ([]claims.ShopRole, error)

	// CRUD: used by the staff handler
	CreateWithUser(ctx context.Context, shopID int, user *models.User, roleID int, profileImage *models.Image) (*models.Staff, error)
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) ([]*models.Staff, error)
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.StaffFilters) (int, error)
	GetByID(ctx context.Context, staffID, shopID int) (*models.Staff, error)
	Update(ctx context.Context, staffID, shopID int, user *models.User, roleID int, profileImage *models.Image) (*models.Staff, error)
	Delete(ctx context.Context, staffID, shopID int) error
	ToggleStatus(ctx context.Context, staffID, shopID int) (*models.Staff, error)
}
