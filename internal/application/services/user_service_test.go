package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func init() {
	logs.Init()
}

func TestUserService_GetByEmail(t *testing.T) {
	t.Run("when user exists then returns user successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "user@example.com"
		expectedUser := &models.User{
			ID:    1,
			Email: email,
		}

		userRepoMock := mocks.NewUserRepository(t)
		authServiceMock := mocks.NewAuthService(t)

		userRepoMock.EXPECT().GetByEmail(ctx, email).Return(expectedUser, nil)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.GetByEmail(ctx, email)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("when user does not exist then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		email := "notfound@example.com"
		expectedError := stdErrors.New(errors.UserNotFound)

		userRepoMock := new(mocks.UserRepository)
		authServiceMock := new(mocks.AuthService)

		userRepoMock.EXPECT().GetByEmail(ctx, email).Return(nil, expectedError)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.GetByEmail(ctx, email)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, user)
	})
}

func TestUserService_ValidateCredentials(t *testing.T) {
	t.Run("when password is valid then returns user successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			ID:       1,
			Email:    "user@example.com",
			Password: "password123",
		}
		hashedPassword := "hashedpassword"

		userRepoMock := new(mocks.UserRepository)
		authServiceMock := new(mocks.AuthService)

		authServiceMock.EXPECT().ComparePassword(ctx, inputUser.Password, hashedPassword).Return(nil)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.ValidateCredentials(ctx, inputUser, hashedPassword)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, inputUser, user)
	})

	t.Run("when password is invalid then returns unauthorized error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			ID:       1,
			Email:    "user@example.com",
			Password: "wrongpassword",
		}
		hashedPassword := "hashedpassword"
		compareError := stdErrors.New("password mismatch")

		userRepoMock := new(mocks.UserRepository)
		authServiceMock := new(mocks.AuthService)

		authServiceMock.EXPECT().ComparePassword(ctx, inputUser.Password, hashedPassword).Return(compareError)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.ValidateCredentials(ctx, inputUser, hashedPassword)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errors.InvalidUserCredentials)
		assert.Nil(t, user)
	})
}

func TestUserService_Create(t *testing.T) {
	t.Run("when user creation is successful then returns created user", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			Email:    "newuser@example.com",
			Password: "password123",
		}
		createdUser := &models.User{
			ID:       1,
			Email:    "newuser@example.com",
			Password: "password123",
		}

		userRepoMock := new(mocks.UserRepository)
		authServiceMock := new(mocks.AuthService)

		userRepoMock.EXPECT().Create(ctx, inputUser).Return(createdUser, nil)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.Create(ctx, inputUser)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, createdUser, user)
	})

	t.Run("when user creation fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			Email:    "existing@example.com",
			Password: "password123",
		}
		expectedError := stdErrors.New(errors.UserAlreadyExists)

		userRepoMock := new(mocks.UserRepository)
		authServiceMock := new(mocks.AuthService)

		userRepoMock.EXPECT().Create(ctx, inputUser).Return(nil, expectedError)

		service := NewUserService(userRepoMock, authServiceMock)

		// Act
		user, err := service.Create(ctx, inputUser)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, user)
	})
}
