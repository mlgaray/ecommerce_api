package errors

const (
	// User related error messages
	UserNotFound           = "user_not_found"
	UserAlreadyExists      = "user_already_exists"
	InvalidUserCredentials = "invalid_credentials"

	// Shop related error messages
	ShopNotFound      = "shop_not_found"
	ShopAlreadyExists = "shop_already_exists"

	// Product related error messages
	ProductNotFound                               = "product_not_found"
	ProductAlreadyExists                          = "product_already_exists"
	ProductPriceMustBePositive                    = "product_price_must_be_positive"
	ProductStockCannotBeNegative                  = "product_stock_cannot_be_negative"
	ProductMinimumStockCannotBeNegative           = "product_minimum_stock_cannot_be_negative"
	ProductMinimumStockCannotBeGreaterThanStock   = "product_minimum_stock_cannot_be_greater_than_stock"
	MinimumStockRequiresStock                     = "minimum_stock_requires_stock"
	PromotionalProductRequiresPromotionalPrice    = "promotional_product_requires_promotional_price"
	PromotionalPriceMustBeLowerThanRegularPrice   = "promotional_price_must_be_lower_than_regular_price"
	PromotionalPriceMustBePositiveWhenPromotional = "promotional_price_must_be_positive_when_promotional"
	QuantityMustBePositive                        = "quantity_must_be_positive"
	InsufficientStock                             = "insufficient_stock"

	// Category related error messages
	CategoryNotFound            = "category_not_found"
	CategoryNameIsRequired      = "category_name_is_required"
	CategoryAlreadyExistsInShop = "category_already_exists_in_shop"
	CategoryImageIsRequired     = "category_image_is_required"
	CategoryHasProducts         = "category_has_products"

	// Authentication related error messages
	TokenExpired            = "token_expired"
	TokenInvalid            = "token_invalid"
	TokenGenerationFailed   = "token_generation_failed"
	TokenCannotBeEmpty      = "token_cannot_be_empty"
	UnexpectedSigningMethod = "unexpected_signing_method"
	CouldNotParseToken      = "could_not_parse_token"

	// Validation error messages
	InvalidInput           = "invalid_input"
	PasswordsCannotBeEmpty = "passwords_cannot_be_empty"

	// ProductFilters related error messages
	ShopIDIsRequired                    = "shop_id_is_required"
	CursorCannotBeNegative              = "cursor_cannot_be_negative"
	MinPriceCannotBeNegative            = "min_price_cannot_be_negative"
	MaxPriceCannotBeNegative            = "max_price_cannot_be_negative"
	MinPriceCannotBeGreaterThanMaxPrice = "min_price_cannot_be_greater_than_max_price"
	InvalidSortField                    = "invalid_sort_field"
	InvalidSortOrder                    = "invalid_sort_order"

	// Authorization error messages
	Forbidden = "forbidden"

	// Payment Method related error messages
	PaymentMethodNotFound          = "payment_method_not_found"
	PaymentMethodIDIsRequired      = "payment_method_id_is_required"
	ShopPaymentMethodNotFound      = "shop_payment_method_not_found"
	ShopPaymentMethodAlreadyExists = "shop_payment_method_already_exists"

	// Transfer config error messages
	TransferConfigurationRequired       = "transfer_configuration_required"
	TransferCBUIsRequired               = "transfer_cbu_is_required"
	TransferCBUInvalidLength            = "transfer_cbu_invalid_length"
	TransferCUILIsRequired              = "transfer_cuil_is_required"
	TransferAccountHolderNameIsRequired = "transfer_account_holder_name_is_required"

	// MercadoPago config error messages
	MercadoPagoConfigurationRequired = "mercadopago_configuration_required"
	MercadoPagoAccessTokenIsRequired = "mercadopago_access_token_is_required"
	MercadoPagoPublicKeyIsRequired   = "mercadopago_public_key_is_required"

	// Delivery Method related error messages
	DeliveryMethodNotFound          = "delivery_method_not_found"
	DeliveryMethodIDIsRequired      = "delivery_method_id_is_required"
	ShopDeliveryMethodNotFound      = "shop_delivery_method_not_found"
	ShopDeliveryMethodAlreadyExists = "shop_delivery_method_already_exists"

	// Delivery zone error messages
	DeliveryZoneNameIsRequired       = "delivery_zone_name_is_required"
	DeliveryZoneCostCannotBeNegative = "delivery_zone_cost_cannot_be_negative"

	// Pickup config error messages
	PickupConfigurationRequired = "pickup_configuration_required"
	PickupAddressIsRequired     = "pickup_address_is_required"
	PickupCityIsRequired        = "pickup_city_is_required"

	// Operating schedule error messages
	InvalidDayOfWeek   = "invalid_day_of_week"
	OpenTimeRequired   = "open_time_required"
	CloseTimeRequired  = "close_time_required"
	InvalidTimeFormat  = "invalid_time_format"
	InvalidTimeRange   = "close_time_must_be_after_open_time"
	ShopClosed         = "shop_is_closed"
	ShopHasNoSchedules = "shop_has_no_operating_schedules_configured"
)
