package models

// PaymentMethodCode represents the payment method codes
type PaymentMethodCode string

const (
	PaymentMethodTransfer    PaymentMethodCode = "transfer"
	PaymentMethodMercadoPago PaymentMethodCode = "mercadopago"
	PaymentMethodCash        PaymentMethodCode = "cash"
)

// PaymentMethod represents a payment method in the catalog.
// When loaded in the context of a Shop, IsActive represents whether the shop has this method enabled,
// and the config fields (TransferConfig, MercadoPagoConfig) contain shop-specific configuration.
type PaymentMethod struct {
	ID          int               `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Code        PaymentMethodCode `json:"code,omitempty"`
	Description string            `json:"description,omitempty"`
	IsActive    bool              `json:"is_active"`

	// Shop-specific fields (only populated when loaded from Shop context)
	TransferConfig    *TransferConfig    `json:"transfer_config,omitempty"`
	MercadoPagoConfig *MercadoPagoConfig `json:"mercadopago_config,omitempty"`
}
