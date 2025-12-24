package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// TransferConfig represents bank transfer configuration for a shop
type TransferConfig struct {
	ID                  int       `json:"id,omitempty"`
	ShopPaymentMethodID int       `json:"shop_payment_method_id,omitempty"`
	CBU                 string    `json:"cbu,omitempty"`
	CUIL                string    `json:"cuil,omitempty"`
	Alias               string    `json:"alias,omitempty"`
	AccountHolderName   string    `json:"account_holder_name,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
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
	if t.AccountHolderName == "" {
		return &errors.ValidationError{
			Message: errors.TransferAccountHolderNameIsRequired,
		}
	}
	return nil
}
