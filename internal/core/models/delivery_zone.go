package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// DeliveryZone represents a delivery zone with its cost
type DeliveryZone struct {
	ID                   int       `json:"id,omitempty"`
	ShopDeliveryMethodID int       `json:"shop_delivery_method_id,omitempty"`
	Name                 string    `json:"name,omitempty"`
	Cost                 float64   `json:"cost"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
}

// Validate validates business rules for DeliveryZone
func (d *DeliveryZone) Validate() error {
	if d.Name == "" {
		return &errors.ValidationError{
			Message: errors.DeliveryZoneNameIsRequired,
		}
	}
	if d.Cost < 0 {
		return &errors.ValidationError{
			Message: errors.DeliveryZoneCostCannotBeNegative,
		}
	}
	return nil
}
