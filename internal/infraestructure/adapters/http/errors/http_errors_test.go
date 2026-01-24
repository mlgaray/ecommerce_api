package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BadRequestError Tests
// =============================================================================

func TestBadRequestError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &BadRequestError{Message: "invalid_input"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "invalid_input", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &BadRequestError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &BadRequestError{Message: "bad_request"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "bad_request", err.Error())
	})
}

// =============================================================================
// UnauthorizedError Tests
// =============================================================================

func TestUnauthorizedError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &UnauthorizedError{Message: "token_expired"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "token_expired", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &UnauthorizedError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &UnauthorizedError{Message: "unauthorized"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "unauthorized", err.Error())
	})
}

// =============================================================================
// ForbiddenError Tests
// =============================================================================

func TestForbiddenError_Error(t *testing.T) {
	t.Run("when message is set then returns message", func(t *testing.T) {
		// Arrange
		err := &ForbiddenError{Message: "access_denied"}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "access_denied", result)
	})

	t.Run("when message is empty then returns empty string", func(t *testing.T) {
		// Arrange
		err := &ForbiddenError{Message: ""}

		// Act
		result := err.Error()

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("implements error interface", func(t *testing.T) {
		// Arrange
		var err error = &ForbiddenError{Message: "forbidden"}

		// Assert
		assert.NotNil(t, err)
		assert.Equal(t, "forbidden", err.Error())
	})
}

// =============================================================================
// Type Assertion Tests
// =============================================================================

func TestHTTPErrorTypeAssertions(t *testing.T) {
	t.Run("BadRequestError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &BadRequestError{Message: "test"}

		// Act
		var badReqErr *BadRequestError
		ok := false
		if e, isType := err.(*BadRequestError); isType {
			badReqErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", badReqErr.Message)
	})

	t.Run("UnauthorizedError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &UnauthorizedError{Message: "test"}

		// Act
		var unauthErr *UnauthorizedError
		ok := false
		if e, isType := err.(*UnauthorizedError); isType {
			unauthErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", unauthErr.Message)
	})

	t.Run("ForbiddenError can be type asserted", func(t *testing.T) {
		// Arrange
		var err error = &ForbiddenError{Message: "test"}

		// Act
		var forbiddenErr *ForbiddenError
		ok := false
		if e, isType := err.(*ForbiddenError); isType {
			forbiddenErr = e
			ok = true
		}

		// Assert
		assert.True(t, ok)
		assert.Equal(t, "test", forbiddenErr.Message)
	})
}
