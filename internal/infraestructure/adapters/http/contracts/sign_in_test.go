package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestSignInRequest_Validate(t *testing.T) {
	t.Run("when request is valid then returns no error", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "user@example.com",
			Password: "password123",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when email is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "",
			Password: "password123",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*errors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "email_is_required", badRequestErr.Message)
	})

	t.Run("when email is only whitespace then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "   ",
			Password: "password123",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*errors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "email_is_required", badRequestErr.Message)
	})

	t.Run("when email format is invalid then returns bad request error", func(t *testing.T) {
		// Arrange
		testCases := []struct {
			name  string
			email string
		}{
			{"missing @ symbol", "userexample.com"},
			{"missing domain", "user@"},
			{"missing local part", "@example.com"},
			{"invalid characters", "user@exa mple.com"},
			{"missing TLD", "user@example"},
			{"double @", "user@@example.com"},
			{"starting with dot", ".user@example.com"},
			{"ending with dot", "user.@example.com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				request := &SignInRequest{
					Email:    tc.email,
					Password: "password123",
				}

				// Act
				err := request.Validate()

				// Assert
				assert.Error(t, err)
				badRequestErr, ok := err.(*errors.BadRequestError)
				assert.True(t, ok)
				assert.Equal(t, "invalid_email_format", badRequestErr.Message)
			})
		}
	})

	t.Run("when password is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "user@example.com",
			Password: "",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*errors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "password_is_required", badRequestErr.Message)
	})

	t.Run("when password is only whitespace then returns bad request error", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "user@example.com",
			Password: "   ",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*errors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "password_is_required", badRequestErr.Message)
	})

	t.Run("when both email and password are empty then returns email error first", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "",
			Password: "",
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*errors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "email_is_required", badRequestErr.Message)
	})
}

func TestSignInRequest_ToUser(t *testing.T) {
	t.Run("when converting to user then returns user with trimmed fields", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "  user@example.com  ",
			Password: "  password123  ",
		}
		expectedUser := &models.User{
			Email:    "user@example.com",
			Password: "password123",
		}

		// Act
		user := request.ToUser()

		// Assert
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Password, user.Password)
		assert.Equal(t, 0, user.ID) // ID should be zero value
	})

	t.Run("when fields have no whitespace then returns user unchanged", func(t *testing.T) {
		// Arrange
		request := &SignInRequest{
			Email:    "user@example.com",
			Password: "password123",
		}
		expectedUser := &models.User{
			Email:    "user@example.com",
			Password: "password123",
		}

		// Act
		user := request.ToUser()

		// Assert
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Password, user.Password)
	})
}
