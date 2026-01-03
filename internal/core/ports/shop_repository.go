package ports

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ShopRepository interface {
	// Create creates a new shop with all payment and delivery methods initialized (is_active = false).
	// This ensures a Shop is always created with its required child entities.
	// userID is the owner of the shop (FK to users table).
	Create(ctx context.Context, userID int, shop *models.Shop) (*models.Shop, error)

	// GetByID returns a shop with all its related entities (cover_image, address, payment_methods,
	// delivery_methods, operating_schedules) loaded. Uses a single query with LEFT JOIN LATERAL.
	// Returns RecordNotFoundError if shop doesn't exist.
	GetByID(ctx context.Context, shopID int) (*models.Shop, error)

	// GetShopsByUserID returns all shops owned by a user.
	// Used during authentication to include shop IDs in JWT token.
	GetShopsByUserID(ctx context.Context, userID int) ([]*models.Shop, error)

	// Operating Schedules (part of Shop aggregate)
	// GetOperatingSchedules returns all operating schedules for a shop, ordered by day_of_week and open_time.
	GetOperatingSchedules(ctx context.Context, shopID int) ([]*models.OperatingSchedule, error)

	// SetOperatingSchedules replaces all operating schedules for a shop.
	// Deletes existing schedules and inserts new ones in a transaction.
	SetOperatingSchedules(ctx context.Context, shopID int, schedules []*models.OperatingSchedule) error

	// IsShopOpen checks if the shop is open at the given time.
	// Returns true if there is an operating schedule that covers the given time.
	IsShopOpen(ctx context.Context, shopID int, checkTime time.Time) (bool, error)

	// Update updates a shop and all its relations using a stored procedure.
	// Returns array of storage_refs of deleted images for Cloudinary cleanup.
	Update(ctx context.Context, shopID int, shop *models.Shop) ([]string, error)
}
