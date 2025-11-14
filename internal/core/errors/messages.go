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
	CategoryNotFound = "category_not_found"

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

	// Authorization error messages
	Forbidden = "forbidden"
)
