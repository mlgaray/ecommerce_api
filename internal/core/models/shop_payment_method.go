package models

import "time"

// ShopPaymentMethod represents a payment method enabled for a shop
type ShopPaymentMethod struct {
	ID              int            `json:"id,omitempty"`
	ShopID          int            `json:"shop_id,omitempty"`
	PaymentMethodID int            `json:"payment_method_id,omitempty"`
	PaymentMethod   *PaymentMethod `json:"payment_method,omitempty"`
	IsActive        bool           `json:"is_active"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`

	// Configurations (loaded based on payment method type)
	TransferConfig    *TransferConfig    `json:"transfer_config,omitempty"`
	MercadoPagoConfig *MercadoPagoConfig `json:"mercadopago_config,omitempty"`
}
