package responses

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// ShopResponse represents a shop in HTTP responses.
type ShopResponse struct {
	ID                 int                              `json:"id,omitempty"`
	Name               string                           `json:"name,omitempty"`
	Slug               string                           `json:"slug,omitempty"`
	Email              string                           `json:"email,omitempty"`
	Phone              string                           `json:"phone,omitempty"`
	Instagram          string                           `json:"instagram,omitempty"`
	PrimaryColor       *string                          `json:"primary_color,omitempty"`
	Images             []*ShopImageResponse             `json:"images,omitempty"`
	Address            *ShopAddressResponse             `json:"address,omitempty"`
	PaymentMethods     []*ShopPaymentMethodResponse     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*ShopDeliveryMethodResponse    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*ShopOperatingScheduleResponse `json:"operating_schedules,omitempty"`
	Timezone           *ShopTimezoneResponse            `json:"timezone,omitempty"`
}

// ShopResponseFromModel converts a domain Shop to a ShopResponse.
func ShopResponseFromModel(shop *models.Shop) *ShopResponse {
	if shop == nil {
		return nil
	}

	return &ShopResponse{
		ID:                 shop.ID,
		Name:               shop.Name,
		Slug:               shop.Slug,
		Email:              shop.Email,
		Phone:              shop.Phone,
		Instagram:          shop.Instagram,
		PrimaryColor:       shop.PrimaryColor,
		Images:             shopImageResponsesFromModels(shop.Images),
		Address:            ShopAddressResponseFromModel(shop.Address),
		PaymentMethods:     shopPaymentMethodResponsesFromModels(shop.PaymentMethods),
		DeliveryMethods:    shopDeliveryMethodResponsesFromModels(shop.DeliveryMethods),
		OperatingSchedules: shopOperatingScheduleResponsesFromModels(shop.OperatingSchedules),
		Timezone:           ShopTimezoneResponseFromModel(shop.Timezone),
	}
}

// ShopImageResponse represents image data in shop responses.
type ShopImageResponse struct {
	ID   int    `json:"id,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

// ShopImageResponseFromModel converts a domain Image to a ShopImageResponse.
func ShopImageResponseFromModel(img *models.Image) *ShopImageResponse {
	if img == nil {
		return nil
	}
	return &ShopImageResponse{
		ID:   img.ID,
		URL:  img.URL,
		Type: img.Type,
	}
}

func shopImageResponsesFromModels(images []*models.Image) []*ShopImageResponse {
	if len(images) == 0 {
		return nil
	}
	response := make([]*ShopImageResponse, len(images))
	for i, img := range images {
		response[i] = ShopImageResponseFromModel(img)
	}
	return response
}

// ShopAddressResponse represents address data in shop responses.
type ShopAddressResponse struct {
	ID      int     `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}

// ShopAddressResponseFromModel converts a domain Address to a ShopAddressResponse.
func ShopAddressResponseFromModel(address *models.Address) *ShopAddressResponse {
	if address == nil {
		return nil
	}
	return &ShopAddressResponse{
		ID:      address.ID,
		Name:    address.Name,
		PlaceID: address.PlaceID,
		Lat:     address.Lat,
		Lng:     address.Lng,
	}
}

// ShopPaymentMethodResponse represents payment method data in shop responses.
type ShopPaymentMethodResponse struct {
	ID                int                            `json:"id,omitempty"`
	Name              string                         `json:"name,omitempty"`
	Code              string                         `json:"code,omitempty"`
	Description       string                         `json:"description,omitempty"`
	IsActive          bool                           `json:"is_active"`
	TransferConfig    *ShopTransferConfigResponse    `json:"transfer_config,omitempty"`
	MercadoPagoConfig *ShopMercadoPagoConfigResponse `json:"mercadopago_config,omitempty"`
}

// ShopPaymentMethodResponseFromModel converts a domain PaymentMethod to a ShopPaymentMethodResponse.
func ShopPaymentMethodResponseFromModel(pm *models.PaymentMethod) *ShopPaymentMethodResponse {
	if pm == nil {
		return nil
	}
	return &ShopPaymentMethodResponse{
		ID:                pm.ID,
		Name:              pm.Name,
		Code:              string(pm.Code),
		Description:       pm.Description,
		IsActive:          pm.IsActive,
		TransferConfig:    ShopTransferConfigResponseFromModel(pm.TransferConfig),
		MercadoPagoConfig: ShopMercadoPagoConfigResponseFromModel(pm.MercadoPagoConfig),
	}
}

func shopPaymentMethodResponsesFromModels(methods []*models.PaymentMethod) []*ShopPaymentMethodResponse {
	if len(methods) == 0 {
		return nil
	}
	response := make([]*ShopPaymentMethodResponse, len(methods))
	for i, pm := range methods {
		response[i] = ShopPaymentMethodResponseFromModel(pm)
	}
	return response
}

// ShopTransferConfigResponse represents transfer configuration in responses.
type ShopTransferConfigResponse struct {
	ID        int    `json:"id,omitempty"`
	CBU       string `json:"cbu,omitempty"`
	CUIL      string `json:"cuil,omitempty"`
	Alias     string `json:"alias,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
}

// ShopTransferConfigResponseFromModel converts a domain TransferConfig to a ShopTransferConfigResponse.
func ShopTransferConfigResponseFromModel(config *models.TransferConfig) *ShopTransferConfigResponse {
	if config == nil {
		return nil
	}
	return &ShopTransferConfigResponse{
		ID:        config.ID,
		CBU:       config.CBU,
		CUIL:      config.CUIL,
		Alias:     config.Alias,
		OwnerName: config.OwnerName,
	}
}

// ShopMercadoPagoConfigResponse represents MercadoPago configuration in responses.
type ShopMercadoPagoConfigResponse struct {
	ID          int    `json:"id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// ShopMercadoPagoConfigResponseFromModel converts a domain MercadoPagoConfig to a ShopMercadoPagoConfigResponse.
func ShopMercadoPagoConfigResponseFromModel(config *models.MercadoPagoConfig) *ShopMercadoPagoConfigResponse {
	if config == nil {
		return nil
	}
	return &ShopMercadoPagoConfigResponse{
		ID:          config.ID,
		AccessToken: config.AccessToken,
		PublicKey:   config.PublicKey,
		UserID:      config.UserID,
	}
}

// ShopDeliveryMethodResponse represents delivery method data in shop responses.
type ShopDeliveryMethodResponse struct {
	ID             int                         `json:"id,omitempty"`
	Name           string                      `json:"name,omitempty"`
	Code           string                      `json:"code,omitempty"`
	Description    string                      `json:"description,omitempty"`
	IsActive       bool                        `json:"is_active"`
	DeliveryConfig *ShopDeliveryConfigResponse `json:"delivery_config,omitempty"`
	DeliveryZones  []*ShopDeliveryZoneResponse `json:"delivery_zones,omitempty"`
	PickupConfig   *ShopPickupConfigResponse   `json:"pickup_config,omitempty"`
}

// ShopDeliveryMethodResponseFromModel converts a domain DeliveryMethod to a ShopDeliveryMethodResponse.
func ShopDeliveryMethodResponseFromModel(dm *models.DeliveryMethod) *ShopDeliveryMethodResponse {
	if dm == nil {
		return nil
	}
	return &ShopDeliveryMethodResponse{
		ID:             dm.ID,
		Name:           dm.Name,
		Code:           string(dm.Code),
		Description:    dm.Description,
		IsActive:       dm.IsActive,
		DeliveryConfig: ShopDeliveryConfigResponseFromModel(dm.DeliveryConfig),
		DeliveryZones:  shopDeliveryZoneResponsesFromModels(dm.DeliveryZones),
		PickupConfig:   ShopPickupConfigResponseFromModel(dm.PickupConfig),
	}
}

func shopDeliveryMethodResponsesFromModels(methods []*models.DeliveryMethod) []*ShopDeliveryMethodResponse {
	if len(methods) == 0 {
		return nil
	}
	response := make([]*ShopDeliveryMethodResponse, len(methods))
	for i, dm := range methods {
		response[i] = ShopDeliveryMethodResponseFromModel(dm)
	}
	return response
}

// ShopDeliveryConfigResponse represents delivery configuration in responses.
type ShopDeliveryConfigResponse struct {
	ID         int      `json:"id,omitempty"`
	FixedPrice *float64 `json:"fixed_price,omitempty"`
}

// ShopDeliveryConfigResponseFromModel converts a domain DeliveryConfig to a ShopDeliveryConfigResponse.
func ShopDeliveryConfigResponseFromModel(config *models.DeliveryConfig) *ShopDeliveryConfigResponse {
	if config == nil {
		return nil
	}
	return &ShopDeliveryConfigResponse{
		ID:         config.ID,
		FixedPrice: config.FixedPrice,
	}
}

// ShopDeliveryZoneResponse represents delivery zone data in responses.
type ShopDeliveryZoneResponse struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Price float64 `json:"price,omitempty"`
}

// ShopDeliveryZoneResponseFromModel converts a domain DeliveryZone to a ShopDeliveryZoneResponse.
func ShopDeliveryZoneResponseFromModel(zone *models.DeliveryZone) *ShopDeliveryZoneResponse {
	if zone == nil {
		return nil
	}
	return &ShopDeliveryZoneResponse{
		ID:    zone.ID,
		Name:  zone.Name,
		Price: zone.Price,
	}
}

func shopDeliveryZoneResponsesFromModels(zones []*models.DeliveryZone) []*ShopDeliveryZoneResponse {
	if len(zones) == 0 {
		return nil
	}
	response := make([]*ShopDeliveryZoneResponse, len(zones))
	for i, zone := range zones {
		response[i] = ShopDeliveryZoneResponseFromModel(zone)
	}
	return response
}

// ShopPickupConfigResponse represents pickup configuration in responses.
type ShopPickupConfigResponse struct {
	ID           int    `json:"id,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	Province     string `json:"province,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Instructions string `json:"instructions,omitempty"`
}

// ShopPickupConfigResponseFromModel converts a domain PickupConfig to a ShopPickupConfigResponse.
func ShopPickupConfigResponseFromModel(config *models.PickupConfig) *ShopPickupConfigResponse {
	if config == nil {
		return nil
	}
	return &ShopPickupConfigResponse{
		ID:           config.ID,
		Address:      config.Address,
		City:         config.City,
		Province:     config.Province,
		PostalCode:   config.PostalCode,
		Instructions: config.Instructions,
	}
}

// ShopOperatingScheduleResponse represents operating schedule data in responses.
type ShopOperatingScheduleResponse struct {
	ID        int    `json:"id,omitempty"`
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time,omitempty"`
	CloseTime string `json:"close_time,omitempty"`
}

// ShopOperatingScheduleResponseFromModel converts a domain OperatingSchedule to a ShopOperatingScheduleResponse.
func ShopOperatingScheduleResponseFromModel(schedule *models.OperatingSchedule) *ShopOperatingScheduleResponse {
	if schedule == nil {
		return nil
	}
	return &ShopOperatingScheduleResponse{
		ID:        schedule.ID,
		DayOfWeek: int(schedule.DayOfWeek),
		OpenTime:  schedule.OpenTime,
		CloseTime: schedule.CloseTime,
	}
}

func shopOperatingScheduleResponsesFromModels(schedules []*models.OperatingSchedule) []*ShopOperatingScheduleResponse {
	if len(schedules) == 0 {
		return nil
	}
	response := make([]*ShopOperatingScheduleResponse, len(schedules))
	for i, sched := range schedules {
		response[i] = ShopOperatingScheduleResponseFromModel(sched)
	}
	return response
}

// ShopTimezoneResponse represents timezone data in responses.
type ShopTimezoneResponse struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	UTCOffset  string `json:"utc_offset,omitempty"`
}

// ShopTimezoneResponseFromModel converts a domain Timezone to a ShopTimezoneResponse.
func ShopTimezoneResponseFromModel(timezone *models.Timezone) *ShopTimezoneResponse {
	if timezone == nil {
		return nil
	}
	return &ShopTimezoneResponse{
		ID:         timezone.ID,
		Name:       timezone.Name,
		Identifier: timezone.Identifier,
		UTCOffset:  timezone.UTCOffset,
	}
}

// GetShopResponse represents the response for a shop retrieval.
type GetShopResponse struct {
	Shop *ShopResponse `json:"shop"`
}
