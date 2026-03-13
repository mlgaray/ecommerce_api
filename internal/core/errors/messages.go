package errors

const (
	// User related error messages
	UserNotFound           = "user_not_found"
	UserAlreadyExists      = "user_already_exists"
	InvalidUserCredentials = "invalid_credentials"

	// Shop related error messages
	ShopNotFound              = "shop_not_found"
	ShopAlreadyExists         = "shop_already_exists"
	InvalidPrimaryColorFormat = "invalid_primary_color_format"

	// Store related error messages
	StoreNotFound = "store_not_found"

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
	TransferConfigurationRequired = "transfer_configuration_required"
	TransferCBUIsRequired         = "transfer_cbu_is_required"
	TransferCBUInvalidLength      = "transfer_cbu_invalid_length"
	TransferCUILIsRequired        = "transfer_cuil_is_required"
	TransferOwnerNameIsRequired   = "transfer_owner_name_is_required"

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
	DeliveryZoneNameIsRequired        = "delivery_zone_name_is_required"
	DeliveryZonePriceCannotBeNegative = "delivery_zone_price_cannot_be_negative"

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

	// Order related error messages
	OrderNotFound               = "order_not_found"
	OrderCannotBeEdited         = "order_cannot_be_edited"
	OrderMustHaveItems          = "order_must_have_items"
	OrderCustomerNameRequired   = "order_customer_name_required"
	OrderItemQuantityInvalid    = "order_item_quantity_must_be_positive"
	OrderItemProductRequired    = "order_item_product_required"
	OrderPaymentMethodRequired  = "order_payment_method_required"
	OrderDeliveryMethodRequired = "order_delivery_method_required"

	// Shipping cost validation errors
	ShippingCostMismatch     = "shipping_cost_does_not_match_delivery_configuration"
	DeliveryZoneNotFound     = "delivery_zone_not_found"
	DeliveryZoneRequired     = "delivery_zone_required_for_zone_based_delivery"
	PickupShouldHaveZeroCost = "pickup_delivery_should_have_zero_shipping_cost"

	// Price validation errors
	UnitPriceMismatch = "unit_price_does_not_match_calculated_price"
	SubtotalMismatch  = "subtotal_does_not_match_calculated_subtotal"
	TotalMismatch     = "total_does_not_match_calculated_total"

	// Order item validation errors
	ProductDataMismatch = "product_data_does_not_match_current_data"
	ProductNotActive    = "product_is_not_active"
	OutOfStock          = "product_is_out_of_stock"

	// Order filters related error messages
	DateFromCannotBeAfterDateTo = "date_from_cannot_be_after_date_to"
	InvalidOrderStatus          = "invalid_order_status"

	// Order state machine error messages
	InvalidOrderTransition      = "invalid_transition"
	OrderConcurrentModification = "concurrent_modification"

	// Coupon related error messages
	CouponNotFound                 = "coupon_not_found"
	CouponNotActive                = "coupon_is_not_active"
	CouponExpired                  = "coupon_is_expired"
	CouponNotYetStarted            = "coupon_has_not_started_yet"
	CouponMinOrderNotMet           = "coupon_min_order_amount_not_met"
	CouponUsageLimitReached        = "coupon_usage_limit_reached"
	CouponPhoneLimitReached        = "coupon_phone_usage_limit_reached"
	CouponCodeIsRequired           = "coupon_code_is_required"
	CouponCodeTooLong              = "coupon_code_exceeds_max_length"
	CouponTypeInvalid              = "coupon_type_must_be_percentage_or_fixed"
	CouponValueMustBePositive      = "coupon_value_must_be_positive"
	CouponPercentageExceedsMax     = "coupon_percentage_cannot_exceed_100"
	CouponMinOrderCannotBeNegative = "coupon_min_order_amount_cannot_be_negative"
	CouponUsageLimitMustBePositive = "coupon_usage_limit_must_be_positive"
	CouponPhoneLimitMustBePositive = "coupon_max_uses_per_phone_must_be_positive"
	CouponExpiresBeforeStarts      = "coupon_expires_at_must_be_after_starts_at"
	CouponAlreadyExists            = "coupon_already_exists_in_shop"
	CouponInvalidCursor            = "invalid_cursor_format"
	OrderAlreadyHasCoupon          = "order_already_has_coupon"
)
