package errors

// RecordNotFoundError represents a domain error when a resource is not found
type RecordNotFoundError struct {
	Message string
}

func (e *RecordNotFoundError) Error() string {
	return e.Message
}

// DuplicateRecordError represents a domain error for constraint violations
// Used when trying to create/update a resource that already exists
type DuplicateRecordError struct {
	Message string
}

func (e *DuplicateRecordError) Error() string {
	return e.Message
}

// ValidationError represents a domain validation error
// Used when business rules or input validation fails
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// AuthenticationError represents failed authentication attempts
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// AuthorizationError represents forbidden access to resources
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// BusinessRuleError represents a violation of business rules
type BusinessRuleError struct {
	Message string
}

func (e *BusinessRuleError) Error() string {
	return e.Message
}
