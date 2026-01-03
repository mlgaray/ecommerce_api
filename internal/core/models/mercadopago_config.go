package models

import (
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// MercadoPagoConfig represents MercadoPago configuration for a shop
type MercadoPagoConfig struct {
	ID          int    `json:"id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	UserID      string `json:"user_id,omitempty"`
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
