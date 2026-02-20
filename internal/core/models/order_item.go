package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// OrderItem representa un item dentro de una orden
type OrderItem struct {
	ID         int       `json:"id,omitempty"`
	OrderID    int       `json:"order_id,omitempty"`
	Product    *Product  `json:"product,omitempty"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
	Notes      string    `json:"notes,omitempty"`
	AddedAt    time.Time `json:"added_at,omitempty"`
}

// Validate validates business rules for the OrderItem domain model
func (oi *OrderItem) Validate() error {
	if oi.Product == nil || oi.Product.ID == 0 {
		return &errors.ValidationError{Message: errors.OrderItemProductRequired}
	}

	if oi.Quantity <= 0 {
		return &errors.ValidationError{Message: errors.OrderItemQuantityInvalid}
	}

	return nil
}
