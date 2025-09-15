package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/mlgaray/ecommerce_api/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

func TestSignInUseCase_Execute(t *testing.T) {
	t.Run("when sign in with valid credentials then returns token successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "user@example.com"
		password := "password123"
		hashedPassword := "hashedpassword"
		expectedToken := "jwt.token.here"

		inputUser := &models.User{
			Email:    email,
			Password: password,
		}

		storedUser := &models.User{
			ID:       1,
			Email:    email,
			Password: hashedPassword,
		}

		authenticatedUser := &models.User{
			ID:    1,
			Email: email,
		}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(authenticatedUser, nil)
		tokenServiceMock.EXPECT().Generate(ctx, authenticatedUser).Return(expectedToken, nil)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("when user not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "notfound@example.com"
		expectedError := errors.New("user_not_found")

		inputUser := &models.User{
			Email:    email,
			Password: "password123",
		}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when credentials are invalid then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "user@example.com"
		password := "wrongpassword"
		hashedPassword := "hashedpassword"
		expectedError := errors.New("invalid credentials")

		inputUser := &models.User{
			Email:    email,
			Password: password,
		}

		storedUser := &models.User{
			ID:       1,
			Email:    email,
			Password: hashedPassword,
		}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when token generation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "user@example.com"
		password := "password123"
		hashedPassword := "hashedpassword"
		expectedError := errors.New("token generation failed")

		inputUser := &models.User{
			Email:    email,
			Password: password,
		}

		storedUser := &models.User{
			ID:       1,
			Email:    email,
			Password: hashedPassword,
		}

		authenticatedUser := &models.User{
			ID:    1,
			Email: email,
		}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(authenticatedUser, nil)
		tokenServiceMock.EXPECT().Generate(ctx, authenticatedUser).Return("", expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})
}
