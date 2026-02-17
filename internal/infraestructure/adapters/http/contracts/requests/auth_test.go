package requests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Sign In - Validate Tests
// =============================================================================

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
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
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
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
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
				badRequestErr, ok := err.(*httpErrors.BadRequestError)
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
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
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
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
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
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "email_is_required", badRequestErr.Message)
	})
}

// =============================================================================
// Sign In - ToUser Tests
// =============================================================================

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

// =============================================================================
// Sign Up - Validate Tests
// =============================================================================

func TestSignUpRequest_Validate(t *testing.T) {
	t.Run("when all fields are valid then returns no error", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "   ",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "invalid-email-format",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john.doe@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
					User: UserRequest{
						Name:     "John",
						LastName: "Doe",
						Email:    email,
						Password: "SecurePassword123!",
						Phone:    "+1234567890",
					},
					Shop: SignUpShopRequest{
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

	t.Run("when valid email with plus sign then passes", func(t *testing.T) {
		// Arrange
		request := SignUpRequest{
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "user+tag@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
			Shop: SignUpShopRequest{
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
					User: UserRequest{
						Name:     "John",
						LastName: "Doe",
						Email:    email,
						Password: "SecurePassword123!",
						Phone:    "+1234567890",
					},
					Shop: SignUpShopRequest{
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

// =============================================================================
// Sign Up - ToUserModel Tests
// =============================================================================

func TestSignUpRequest_ToUserModel(t *testing.T) {
	t.Run("when converting to user model then maps all fields", func(t *testing.T) {
		// Arrange
		request := &SignUpRequest{
			User: UserRequest{
				Name:     "John",
				LastName: "Doe",
				Email:    "john@example.com",
				Password: "SecurePassword123!",
				Phone:    "+1234567890",
			},
		}

		// Act
		user := request.ToUserModel()

		// Assert
		assert.Equal(t, "John", user.Name)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Equal(t, "SecurePassword123!", user.Password)
		assert.Equal(t, "+1234567890", user.Phone)
		assert.Equal(t, 0, user.ID)
	})

	t.Run("when fields have values then preserves them exactly", func(t *testing.T) {
		// Arrange
		request := &SignUpRequest{
			User: UserRequest{
				Name:     "  Maria  ",
				LastName: "  Garcia  ",
				Email:    "maria@example.com",
				Password: "pass123",
				Phone:    "+54 11 5555-1234",
			},
		}

		// Act
		user := request.ToUserModel()

		// Assert
		assert.Equal(t, "  Maria  ", user.Name)
		assert.Equal(t, "  Garcia  ", user.LastName)
		assert.Equal(t, "maria@example.com", user.Email)
	})
}

// =============================================================================
// Sign Up - ToShopModel Tests
// =============================================================================

func TestSignUpRequest_ToShopModel(t *testing.T) {
	t.Run("when converting to shop model then maps all fields", func(t *testing.T) {
		// Arrange
		request := &SignUpRequest{
			Shop: SignUpShopRequest{
				Name:  "My Shop",
				Slug:  "my-shop",
				Email: "shop@example.com",
				Phone: "+0987654321",
			},
		}

		// Act
		shop := request.ToShopModel()

		// Assert
		assert.Equal(t, "My Shop", shop.Name)
		assert.Equal(t, "my-shop", shop.Slug)
		assert.Equal(t, "shop@example.com", shop.Email)
		assert.Equal(t, "+0987654321", shop.Phone)
		assert.Equal(t, 0, shop.ID)
	})

	t.Run("when shop has all fields then preserves them", func(t *testing.T) {
		// Arrange
		request := &SignUpRequest{
			Shop: SignUpShopRequest{
				Name:  "Tienda de Ropa",
				Slug:  "tienda-de-ropa",
				Email: "tienda@example.com",
				Phone: "+54 11 5555-0000",
			},
		}

		// Act
		shop := request.ToShopModel()

		// Assert
		assert.Equal(t, "Tienda de Ropa", shop.Name)
		assert.Equal(t, "tienda-de-ropa", shop.Slug)
		assert.Nil(t, shop.Images)
		assert.Nil(t, shop.Address)
		assert.Nil(t, shop.PaymentMethods)
	})
}
