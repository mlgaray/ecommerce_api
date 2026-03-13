package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// RemoveOrderCouponUseCase removes the coupon from an order and recalculates totals.
// The stored procedure handles coupon_usages cleanup when coupon_id is NULL.
type RemoveOrderCouponUseCase interface {
	Execute(ctx context.Context, orderID, shopID int) (*models.Order, error)
}
