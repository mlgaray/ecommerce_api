package models

import (
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// TransferConfig represents bank transfer configuration for a shop
type TransferConfig struct {
	ID        int    `json:"id,omitempty"`
	CBU       string `json:"cbu,omitempty"`
	CUIL      string `json:"cuil,omitempty"`
	Alias     string `json:"alias,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
}

// Validate validates business rules for TransferConfig
func (t *TransferConfig) Validate() error {
	if t.CBU == "" {
		return &errors.ValidationError{
			Message: errors.TransferCBUIsRequired,
		}
	}
	if len(t.CBU) != 22 {
		return &errors.ValidationError{
			Message: errors.TransferCBUInvalidLength,
		}
	}
	if t.CUIL == "" {
		return &errors.ValidationError{
			Message: errors.TransferCUILIsRequired,
		}
	}
	if t.OwnerName == "" {
		return &errors.ValidationError{
			Message: errors.TransferOwnerNameIsRequired,
		}
	}
	return nil
}
