package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// OrderEventNotifier emits domain events when orders change.
// Infrastructure layer provides the concrete implementation (e.g., WebSocket hub, Redis pub/sub).
// Notifications are fire-and-forget: failures are logged but do NOT affect the caller.
type OrderEventNotifier interface {
	// NotifyNewOrder notifies listeners that a new order has been created.
	// The shopID is derived from order.Store.ID.
	// Implementations must be safe for concurrent use.
	NotifyNewOrder(ctx context.Context, order *models.Order)
}
