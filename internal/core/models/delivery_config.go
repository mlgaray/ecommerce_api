package models

// DeliveryConfig represents delivery configuration for a shop's delivery method.
// If FixedPrice is set (including 0 for free shipping), it overrides zone-based pricing.
// If FixedPrice is nil, use DeliveryZones for zone-based pricing.
type DeliveryConfig struct {
	ID         int      `json:"id,omitempty"`
	FixedPrice *float64 `json:"fixed_price,omitempty"`
}
