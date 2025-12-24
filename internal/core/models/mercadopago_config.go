package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// MercadoPagoConfig represents MercadoPago configuration for a shop
type MercadoPagoConfig struct {
	ID                  int       `json:"id,omitempty"`
	ShopPaymentMethodID int       `json:"shop_payment_method_id,omitempty"`
	AccessToken         string    `json:"-"` // Never expose in JSON responses
	PublicKey           string    `json:"public_key,omitempty"`
	UserID              string    `json:"user_id,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}

// Validate validates business rules for MercadoPagoConfig
func (m *MercadoPagoConfig) Validate() error {
	if m.AccessToken == "" {
		return &errors.ValidationError{
			Message: errors.MercadoPagoAccessTokenIsRequired,
		}
	}
	if m.PublicKey == "" {
		return &errors.ValidationError{
			Message: errors.MercadoPagoPublicKeyIsRequired,
		}
	}
	return nil
}
