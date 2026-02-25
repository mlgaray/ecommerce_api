package stubs

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// OrderEventNotifier is a stub implementation for integration tests.
// Notifications are fire-and-forget so this is a no-op.
type OrderEventNotifier struct{}

// NewOrderEventNotifier creates a new stub OrderEventNotifier.
func NewOrderEventNotifier() *OrderEventNotifier {
	return &OrderEventNotifier{}
}

// NotifyNewOrder is a no-op for tests.
func (n *OrderEventNotifier) NotifyNewOrder(ctx context.Context, order *models.Order) {
	// No-op: notifications are irrelevant for integration tests
}
