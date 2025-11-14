package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

func TestSignUpRequest_Validate(t *testing.T) {
	t.Run("when all fields are valid then returns no error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when user name is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_name_is_required", badRequestErr.Message)
	})

	t.Run("when user name is only whitespace then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "   ",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_name_is_required", badRequestErr.Message)
	})

	t.Run("when user last name is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_last_name_is_required", badRequestErr.Message)
	})

	t.Run("when user email is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_email_is_required", badRequestErr.Message)
	})

	t.Run("when user email format is invalid then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "invalid-email-format",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "invalid_email_format", badRequestErr.Message)
	})

	t.Run("when user phone is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_phone_is_required", badRequestErr.Message)
	})

	t.Run("when user password is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "user_password_is_required", badRequestErr.Message)
	})

	t.Run("when shop name is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "shop_name_is_required", badRequestErr.Message)
	})

	t.Run("when shop slug is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "shop_slug_is_required", badRequestErr.Message)
	})

	t.Run("when shop email is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "",
				Phone: "+0987654321",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "shop_email_is_required", badRequestErr.Message)
	})

	t.Run("when shop phone is empty then returns bad request error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: models.User{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: models.Shop{
				Name:  "John's Shop",
				Slug:  "johns-shop",
				Email: "shop@example.com",
				Phone: "",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "shop_phone_is_required", badRequestErr.Message)
	})

	t.Run("when email format is valid with various patterns then returns no error", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"test.email@domain.org",
			"user123@example.co.uk",
			"user_name@example-domain.com",
			"user+tag@example.com",
			"a@b.co",
		}

		for _, email := range validEmails {
			t.Run("email: "+email, func(t *testing.T) {
				// Arrange
				request := SignUpRequest{
					User: models.User{
						Name:     "John",
						LastName: "Doe",
						Email:    email,
						Password: "SecurePassword123!",
						Phone:    "+1234567890",
					},
					Shop: models.Shop{
						Name:  "John's Shop",
						Slug:  "johns-shop",
						Email: "shop@example.com",
						Phone: "+0987654321",
					},
				}

				// Act
				err := request.Validate()

				// Assert
				assert.NoError(t, err)
			})
		}
	})

	t.Run("when email format is invalid with various patterns then returns bad request error", func(t *testing.T) {
		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"user@",
			"user@.com",
			"user.example.com",
			"user @example.com",
			"user@example",
			"",
			"user@@example.com",
		}

		for _, email := range invalidEmails {
			t.Run("email: "+email, func(t *testing.T) {
				// Arrange
				request := SignUpRequest{
					User: models.User{
						Name:     "John",
						LastName: "Doe",
						Email:    email,
						Password: "SecurePassword123!",
						Phone:    "+1234567890",
					},
					Shop: models.Shop{
						Name:  "John's Shop",
						Slug:  "johns-shop",
						Email: "shop@example.com",
						Phone: "+0987654321",
					},
				}

				// Act
				err := request.Validate()

				// Assert
				if email == "" {
					// Empty email should trigger user_email_is_required first
					assert.Error(t, err)
					badRequestErr, ok := err.(*httpErrors.BadRequestError)
					assert.True(t, ok)
					assert.Equal(t, "user_email_is_required", badRequestErr.Message)
				} else {
					// Invalid format should trigger invalid_email_format
					assert.Error(t, err)
					badRequestErr, ok := err.(*httpErrors.BadRequestError)
					assert.True(t, ok)
					assert.Equal(t, "invalid_email_format", badRequestErr.Message)
				}
			})
		}
	})
}
