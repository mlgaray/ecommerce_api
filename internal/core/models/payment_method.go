package models

import "time"

// PaymentMethodCode represents the payment method codes
type PaymentMethodCode string

const (
	PaymentMethodTransfer    PaymentMethodCode = "transfer"
	PaymentMethodMercadoPago PaymentMethodCode = "mercadopago"
	PaymentMethodCash        PaymentMethodCode = "cash"
)

// PaymentMethod represents a payment method in the catalog
type PaymentMethod struct {
	ID          int               `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Code        PaymentMethodCode `json:"code,omitempty"`
	Description string            `json:"description,omitempty"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
}
