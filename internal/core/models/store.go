package models

// Store represents the public-facing view of a shop for customers.
// Similar to Shop but can have additional calculated fields in the future (e.g., IsOpen).
type Store struct {
	ID                 int                  `json:"id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Slug               string               `json:"slug,omitempty"`
	Email              string               `json:"email,omitempty"`
	Phone              string               `json:"phone,omitempty"`
	Instagram          string               `json:"instagram,omitempty"`
	Images             []*Image             `json:"images,omitempty"`
	Address            *Address             `json:"address,omitempty"`
	PaymentMethods     []*PaymentMethod     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*DeliveryMethod    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*OperatingSchedule `json:"operating_schedules,omitempty"`
	// Future: IsOpen bool `json:"is_open"` (requires timezone support)
}

// NewStoreFromShop creates a Store from a Shop model.
func NewStoreFromShop(shop *Shop) *Store {
	if shop == nil {
		return nil
	}

	return &Store{
		ID:                 shop.ID,
		Name:               shop.Name,
		Slug:               shop.Slug,
		Email:              shop.Email,
		Phone:              shop.Phone,
		Instagram:          shop.Instagram,
		Images:             shop.Images,
		Address:            shop.Address,
		PaymentMethods:     shop.PaymentMethods,
		DeliveryMethods:    shop.DeliveryMethods,
		OperatingSchedules: shop.OperatingSchedules,
	}
}
