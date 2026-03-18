package responses

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// Reuse shop response subtypes for store responses (same shape).
type StoreImageResponse = ShopImageResponse
type StoreAddressResponse = ShopAddressResponse
type StorePaymentMethodResponse = ShopPaymentMethodResponse
type StoreTransferConfigResponse = ShopTransferConfigResponse
type StoreMercadoPagoConfigResponse = ShopMercadoPagoConfigResponse
type StoreDeliveryMethodResponse = ShopDeliveryMethodResponse
type StoreDeliveryConfigResponse = ShopDeliveryConfigResponse
type StoreDeliveryZoneResponse = ShopDeliveryZoneResponse
type StorePickupConfigResponse = ShopPickupConfigResponse
type StoreOperatingScheduleResponse = ShopOperatingScheduleResponse
type StoreTimezoneResponse = ShopTimezoneResponse

// StoreResponse represents a store in HTTP responses.
type StoreResponse struct {
	ID                 int                               `json:"id,omitempty"`
	Name               string                            `json:"name,omitempty"`
	Slug               string                            `json:"slug,omitempty"`
	Email              string                            `json:"email,omitempty"`
	Phone              string                            `json:"phone,omitempty"`
	Instagram          string                            `json:"instagram,omitempty"`
	PrimaryColor       *string                           `json:"primary_color,omitempty"`
	Images             []*StoreImageResponse             `json:"images,omitempty"`
	Address            *StoreAddressResponse             `json:"address,omitempty"`
	PaymentMethods     []*StorePaymentMethodResponse     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*StoreDeliveryMethodResponse    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*StoreOperatingScheduleResponse `json:"operating_schedules,omitempty"`
	Timezone           *StoreTimezoneResponse            `json:"timezone,omitempty"`
	IsOpen             bool                              `json:"is_open"`
}

// StoreResponseFromModel converts a domain Store to a StoreResponse.
func StoreResponseFromModel(store *models.Store) *StoreResponse {
	if store == nil {
		return nil
	}

	return &StoreResponse{
		ID:                 store.ID,
		Name:               store.Name,
		Slug:               store.Slug,
		Email:              store.Email,
		Phone:              store.Phone,
		Instagram:          store.Instagram,
		PrimaryColor:       store.PrimaryColor,
		Images:             shopImageResponsesFromModels(store.Images),
		Address:            ShopAddressResponseFromModel(store.Address),
		PaymentMethods:     shopPaymentMethodResponsesFromModels(store.PaymentMethods),
		DeliveryMethods:    shopDeliveryMethodResponsesFromModels(store.DeliveryMethods),
		OperatingSchedules: shopOperatingScheduleResponsesFromModels(store.OperatingSchedules),
		Timezone:           ShopTimezoneResponseFromModel(store.Timezone),
		IsOpen:             store.IsOpen,
	}
}

// GetStoreResponse represents the response for a store retrieval.
type GetStoreResponse struct {
	Store *StoreResponse `json:"store"`
}

// CheckSlugAvailabilityResponse represents the response for a slug availability check.
type CheckSlugAvailabilityResponse struct {
	Available bool `json:"available"`
}
