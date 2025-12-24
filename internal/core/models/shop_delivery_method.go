package models

import "time"

// ShopDeliveryMethod represents a delivery method enabled for a shop
type ShopDeliveryMethod struct {
	ID               int             `json:"id,omitempty"`
	ShopID           int             `json:"shop_id,omitempty"`
	DeliveryMethodID int             `json:"delivery_method_id,omitempty"`
	DeliveryMethod   *DeliveryMethod `json:"delivery_method,omitempty"`
	IsActive         bool            `json:"is_active"`
	CreatedAt        time.Time       `json:"created_at,omitempty"`

	// Configurations (loaded based on delivery method type)
	DeliveryZones []*DeliveryZone `json:"delivery_zones,omitempty"`
	PickupConfig  *PickupConfig   `json:"pickup_config,omitempty"`
}
