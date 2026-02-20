package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// RecordNotFoundError Tests
// =============================================================================

func TestRecordNotFoundError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &RecordNotFoundError{Message: "user_not_found"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "user_not_found", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &RecordNotFoundError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &RecordNotFoundError{Message: "not_found"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "not_found", err.Error())
	})
}

// =============================================================================
// DuplicateRecordError Tests
// =============================================================================

func TestDuplicateRecordError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &DuplicateRecordError{Message: "user_already_exists"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "user_already_exists", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &DuplicateRecordError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &DuplicateRecordError{Message: "duplicate"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "duplicate", err.Error())
	})
}

// =============================================================================
// ValidationError Tests
// =============================================================================

func TestValidationError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &ValidationError{Message: "invalid_email_format"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "invalid_email_format", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &ValidationError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &ValidationError{Message: "validation_failed"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "validation_failed", err.Error())
	})
}

// =============================================================================
// AuthenticationError Tests
// =============================================================================

func TestAuthenticationError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &AuthenticationError{Message: "invalid_credentials"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "invalid_credentials", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &AuthenticationError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &AuthenticationError{Message: "auth_failed"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "auth_failed", err.Error())
	})
}

// =============================================================================
// AuthorizationError Tests
// =============================================================================

func TestAuthorizationError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &AuthorizationError{Message: "forbidden"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "forbidden", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &AuthorizationError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &AuthorizationError{Message: "access_denied"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "access_denied", err.Error())
	})
}

// =============================================================================
// BusinessRuleError Tests
// =============================================================================

func TestBusinessRuleError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &BusinessRuleError{Message: "insufficient_stock"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "insufficient_stock", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &BusinessRuleError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &BusinessRuleError{Message: "rule_violation"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "rule_violation", err.Error())
	})
}

// =============================================================================
// ReferentialIntegrityError Tests
// =============================================================================

func TestReferentialIntegrityError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &ReferentialIntegrityError{Message: "category_has_products"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "category_has_products", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &ReferentialIntegrityError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &ReferentialIntegrityError{Message: "fk_violation"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "fk_violation", err.Error())
	})
}

// =============================================================================
// ConcurrentModificationError Tests
// =============================================================================

func TestConcurrentModificationError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &ConcurrentModificationError{Message: "concurrent_modification"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "concurrent_modification", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &ConcurrentModificationError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &ConcurrentModificationError{Message: "conflict"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "conflict", err.Error())
	})
}

// =============================================================================
// Type Assertion Tests
// =============================================================================

func TestErrorTypeAssertions(t *testing.T) {
	t.Run("RecordNotFoundError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &RecordNotFoundError{Message: "test"}

		// Act
		var notFoundErr *RecordNotFoundError
		ok := false
		if e, isType := err.(*RecordNotFoundError); isType {
			notFoundErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", notFoundErr.Message)
	})

	t.Run("DuplicateRecordError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &DuplicateRecordError{Message: "test"}

		// Act
		var dupErr *DuplicateRecordError
		ok := false
		if e, isType := err.(*DuplicateRecordError); isType {
			dupErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", dupErr.Message)
	})

	t.Run("ValidationError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &ValidationError{Message: "test"}

		// Act
		var valErr *ValidationError
		ok := false
		if e, isType := err.(*ValidationError); isType {
			valErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", valErr.Message)
	})

	t.Run("AuthenticationError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &AuthenticationError{Message: "test"}

		// Act
		var authErr *AuthenticationError
		ok := false
		if e, isType := err.(*AuthenticationError); isType {
			authErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", authErr.Message)
	})

	t.Run("AuthorizationError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &AuthorizationError{Message: "test"}

		// Act
		var authzErr *AuthorizationError
		ok := false
		if e, isType := err.(*AuthorizationError); isType {
			authzErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", authzErr.Message)
	})

	t.Run("BusinessRuleError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &BusinessRuleError{Message: "test"}

		// Act
		var bizErr *BusinessRuleError
		ok := false
		if e, isType := err.(*BusinessRuleError); isType {
			bizErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", bizErr.Message)
	})

	t.Run("ReferentialIntegrityError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &ReferentialIntegrityError{Message: "test"}

		// Act
		var refErr *ReferentialIntegrityError
		ok := false
		if e, isType := err.(*ReferentialIntegrityError); isType {
			refErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", refErr.Message)
	})

	t.Run("ConcurrentModificationError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &ConcurrentModificationError{Message: "test"}

		// Act
		var concErr *ConcurrentModificationError
		ok := false
		if e, isType := err.(*ConcurrentModificationError); isType {
			concErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", concErr.Message)
	})
}

// =============================================================================
// Usage with Error Constants Tests
// =============================================================================

func TestErrorWithConstants(t *testing.T) {
	t.Run("RecordNotFoundError with UserNotFound constant", func(t *testing.T) {
		// Arrange
		err := &RecordNotFoundError{Message: UserNotFound}

		// Assert
		assert.Equal(t, "user_not_found", err.Error())
	})

	t.Run("DuplicateRecordError with UserAlreadyExists constant", func(t *testing.T) {
		// Arrange
		err := &DuplicateRecordError{Message: UserAlreadyExists}

		// Assert
		assert.Equal(t, "user_already_exists", err.Error())
	})

	t.Run("ValidationError with InvalidSortField constant", func(t *testing.T) {
		// Arrange
		err := &ValidationError{Message: InvalidSortField}

		// Assert
		assert.Equal(t, "invalid_sort_field", err.Error())
	})

	t.Run("AuthenticationError with InvalidUserCredentials constant", func(t *testing.T) {
		// Arrange
		err := &AuthenticationError{Message: InvalidUserCredentials}

		// Assert
		assert.Equal(t, "invalid_credentials", err.Error())
	})

	t.Run("AuthorizationError with Forbidden constant", func(t *testing.T) {
		// Arrange
		err := &AuthorizationError{Message: Forbidden}

		// Assert
		assert.Equal(t, "forbidden", err.Error())
	})

	t.Run("ReferentialIntegrityError with CategoryHasProducts constant", func(t *testing.T) {
		// Arrange
		err := &ReferentialIntegrityError{Message: CategoryHasProducts}

		// Assert
		assert.Equal(t, "category_has_products", err.Error())
	})

	t.Run("ConcurrentModificationError with OrderConcurrentModification constant", func(t *testing.T) {
		// Arrange
		err := &ConcurrentModificationError{Message: OrderConcurrentModification}

		// Assert
		assert.Equal(t, "concurrent_modification", err.Error())
	})
}
