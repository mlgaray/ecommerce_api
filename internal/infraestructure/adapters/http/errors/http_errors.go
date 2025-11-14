package errors

// BadRequestError represents HTTP 400 errors (invalid input format, missing fields, etc.)
// This is an HTTP concern, not a domain concern
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}
