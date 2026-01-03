package models

import (
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// DeliveryZone represents a delivery zone with its price
type DeliveryZone struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Price float64 `json:"price"`
}

// Validate validates business rules for DeliveryZone
func (d *DeliveryZone) Validate() error {
	if d.Name == "" {
		return &errors.ValidationError{
			Message: errors.DeliveryZoneNameIsRequired,
		}
	}
	if d.Price < 0 {
		return &errors.ValidationError{
			Message: errors.DeliveryZonePriceCannotBeNegative,
		}
	}
	return nil
}
