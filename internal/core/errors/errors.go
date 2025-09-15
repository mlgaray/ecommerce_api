package errors

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// InternalServiceError represents an internal service error
type InternalServiceError struct {
	Message string
}

func (e *InternalServiceError) Error() string {
	return e.Message
}

// ForbiddenError represents a forbidden access error
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// UnauthorizedError represents an authentication error
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// BadRequestError represents a validation or bad request error
type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// ConflictError represents a conflict error (duplicate resource)
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}
