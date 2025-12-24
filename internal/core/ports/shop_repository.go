package ports

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ShopRepository interface {
	// Create creates a new shop with all payment and delivery methods initialized (is_active = false).
	// This ensures a Shop is always created with its required child entities.
	Create(ctx context.Context, shop *models.Shop) (*models.Shop, error)

	// GetShopsByUserID returns all shops owned by a user.
	// Used during authentication to include shop IDs in JWT token.
	GetShopsByUserID(ctx context.Context, userID int) ([]*models.Shop, error)

	// Payment Methods (read-only, part of Shop aggregate)
	GetPaymentMethods(ctx context.Context, shopID int) ([]*models.ShopPaymentMethod, error)

	// Delivery Methods (read-only, part of Shop aggregate)
	GetDeliveryMethods(ctx context.Context, shopID int) ([]*models.ShopDeliveryMethod, error)

	// Operating Schedules (part of Shop aggregate)
	// GetOperatingSchedules returns all operating schedules for a shop, ordered by day_of_week and open_time.
	GetOperatingSchedules(ctx context.Context, shopID int) ([]*models.OperatingSchedule, error)

	// SetOperatingSchedules replaces all operating schedules for a shop.
	// Deletes existing schedules and inserts new ones in a transaction.
	SetOperatingSchedules(ctx context.Context, shopID int, schedules []*models.OperatingSchedule) error

	// IsShopOpen checks if the shop is open at the given time.
	// Returns true if there is an operating schedule that covers the given time.
	IsShopOpen(ctx context.Context, shopID int, checkTime time.Time) (bool, error)
}
