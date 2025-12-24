package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// PickupConfig represents pickup configuration for a shop
type PickupConfig struct {
	ID                   int       `json:"id,omitempty"`
	ShopDeliveryMethodID int       `json:"shop_delivery_method_id,omitempty"`
	Address              string    `json:"address,omitempty"`
	City                 string    `json:"city,omitempty"`
	Province             string    `json:"province,omitempty"`
	PostalCode           string    `json:"postal_code,omitempty"`
	Instructions         string    `json:"instructions,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
}

// Validate validates business rules for PickupConfig
func (p *PickupConfig) Validate() error {
	if p.Address == "" {
		return &errors.ValidationError{
			Message: errors.PickupAddressIsRequired,
		}
	}
	if p.City == "" {
		return &errors.ValidationError{
			Message: errors.PickupCityIsRequired,
		}
	}
	return nil
}
