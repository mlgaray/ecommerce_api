package errors

// BadRequestError represents HTTP 400 errors (invalid input format, missing fields, etc.)
// This is an HTTP concern, not a domain concern
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// UnauthorizedError represents HTTP 401 errors (missing or invalid authentication)
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}
