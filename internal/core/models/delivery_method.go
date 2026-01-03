package models

// DeliveryMethodCode represents the delivery method codes
type DeliveryMethodCode string

const (
	DeliveryMethodDelivery DeliveryMethodCode = "delivery"
	DeliveryMethodPickup   DeliveryMethodCode = "pickup"
)

// DeliveryMethod represents a delivery method in the catalog.
// When loaded in the context of a Shop, IsActive represents whether the shop has this method enabled,
// and the config fields contain shop-specific configuration.
type DeliveryMethod struct {
	ID          int                `json:"id,omitempty"`
	Name        string             `json:"name,omitempty"`
	Code        DeliveryMethodCode `json:"code,omitempty"`
	Description string             `json:"description,omitempty"`
	IsActive    bool               `json:"is_active"`

	// Shop-specific fields (only populated when loaded from Shop context)
	// For "delivery" type: DeliveryConfig (fixed price) or DeliveryZones (zone-based)
	// For "pickup" type: PickupConfig
	DeliveryConfig *DeliveryConfig `json:"delivery_config,omitempty"`
	DeliveryZones  []*DeliveryZone `json:"delivery_zones,omitempty"`
	PickupConfig   *PickupConfig   `json:"pickup_config,omitempty"`
}
