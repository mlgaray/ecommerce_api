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

		userShops := []*models.Shop{
			{ID: 10, Name: "Test Shop"},
		}
		expectedShopIDs := []int{10}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		shopRepositoryMock := new(mocks.ShopRepository)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(inputUser, nil)
		shopRepositoryMock.EXPECT().GetShopsByUserID(ctx, storedUser.ID).Return(userShops, nil)
		tokenServiceMock.EXPECT().Generate(ctx, storedUser, expectedShopIDs).Return(expectedToken, nil)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, shopRepositoryMock)

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
		shopRepositoryMock := new(mocks.ShopRepository)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, shopRepositoryMock)

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
		shopRepositoryMock := new(mocks.ShopRepository)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, shopRepositoryMock)

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

		userShops := []*models.Shop{
			{ID: 10, Name: "Test Shop"},
		}
		expectedShopIDs := []int{10}

		userServiceMock := new(mocks.UserService)
		tokenServiceMock := new(mocks.TokenService)
		shopRepositoryMock := new(mocks.ShopRepository)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(inputUser, nil)
		shopRepositoryMock.EXPECT().GetShopsByUserID(ctx, storedUser.ID).Return(userShops, nil)
		tokenServiceMock.EXPECT().Generate(ctx, storedUser, expectedShopIDs).Return("", expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, shopRepositoryMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})

	t.Run("when get shops fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "user@example.com"
		password := "password123"
		hashedPassword := "hashedpassword"
		expectedError := errors.New("database error")

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
		shopRepositoryMock := new(mocks.ShopRepository)

		userServiceMock.EXPECT().GetByEmail(ctx, email).Return(storedUser, nil)
		userServiceMock.EXPECT().ValidateCredentials(ctx, inputUser, hashedPassword).Return(inputUser, nil)
		shopRepositoryMock.EXPECT().GetShopsByUserID(ctx, storedUser.ID).Return(nil, expectedError)

		useCase := NewSignInUseCase(userServiceMock, tokenServiceMock, shopRepositoryMock)

		// Act
		token, err := useCase.Execute(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, token)
	})
}
