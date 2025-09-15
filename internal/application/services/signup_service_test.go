package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

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
			ID:       1,
			Email:    "newuser@example.com",
			Password: "password123",
		}

		signupRepoMock := new(mocks.SignupRepository)
		signupRepoMock.EXPECT().CreateUserWithShop(ctx, inputUser, inputShop).Return(expectedUser, nil)

		service := NewSignupService(signupRepoMock)

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
		signupRepoMock.EXPECT().CreateUserWithShop(ctx, inputUser, inputShop).Return(nil, expectedError)

		service := NewSignupService(signupRepoMock)

		// Act
		user, err := service.SignUp(ctx, inputUser, inputShop)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, user)
	})
}
