package errors

const (
	// User related error messages
	UserNotFound           = "user_not_found"
	UserAlreadyExists      = "user_already_exists"
	InvalidUserCredentials = "invalid_credentials"

	// Shop related error messages
	ShopNotFound      = "shop_not_found"
	ShopAlreadyExists = "shop_already_exists"

	// Authentication related error messages
	TokenExpired          = "token_expired"
	TokenInvalid          = "token_invalid"
	TokenGenerationFailed = "token_generation_failed"

	// General error messages
	InternalServerError = "internal_server_error"
	BadRequest          = "bad_request"
	Unauthorized        = "unauthorized"
	Forbidden           = "forbidden"
	Conflict            = "conflict"
	DatabaseError       = "database_error"
)
