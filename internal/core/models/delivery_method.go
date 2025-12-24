package models

import "time"

// DeliveryMethodCode represents the delivery method codes
type DeliveryMethodCode string

const (
	DeliveryMethodDelivery DeliveryMethodCode = "delivery"
	DeliveryMethodPickup   DeliveryMethodCode = "pickup"
)

// DeliveryMethod represents a delivery method in the catalog
type DeliveryMethod struct {
	ID          int                `json:"id,omitempty"`
	Name        string             `json:"name,omitempty"`
	Code        DeliveryMethodCode `json:"code,omitempty"`
	Description string             `json:"description,omitempty"`
	IsActive    bool               `json:"is_active"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
}
