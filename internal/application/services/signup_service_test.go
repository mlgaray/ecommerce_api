package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestSignupService_SignUp(t *testing.T) {
	t.Run("when signup is successful then returns created user", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			Email:    "newuser@example.com",
			Password: "password123",
		}
		inputShop := &models.Shop{
			Name: "Test Shop",
		}
		expectedUser := &models.User{
			ID:    1,
			Email: "newuser@example.com",
		}

		signupRepoMock := new(mocks.SignupRepository)
		authServiceMock := new(mocks.AuthService)

		authServiceMock.EXPECT().HashPassword(ctx, "password123").Return("hashed_password123", nil)
		signupRepoMock.EXPECT().CreateUserWithShop(ctx, mock.AnythingOfType("*models.User"), inputShop).Return(expectedUser, nil)

		service := NewSignupService(signupRepoMock, authServiceMock)

		// Act
		user, err := service.SignUp(ctx, inputUser, inputShop)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("when signup fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		inputUser := &models.User{
			Email:    "existing@example.com",
			Password: "password123",
		}
		inputShop := &models.Shop{
			Name: "Test Shop",
		}
		expectedError := stdErrors.New("user already exists")

		signupRepoMock := mocks.NewSignupRepository(t)
		authServiceMock := new(mocks.AuthService)

		authServiceMock.EXPECT().HashPassword(ctx, "password123").Return("hashed_password123", nil)
		signupRepoMock.EXPECT().CreateUserWithShop(ctx, mock.AnythingOfType("*models.User"), inputShop).Return(nil, expectedError)

		service := NewSignupService(signupRepoMock, authServiceMock)

		// Act
		user, err := service.SignUp(ctx, inputUser, inputShop)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, user)
	})
}
