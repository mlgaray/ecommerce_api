package auth

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestSignUpUseCase_Execute(t *testing.T) {
	t.Run("when sign up with valid data then returns success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		inputUser := &models.User{
			Email:    "user@example.com",
			Password: "password123",
		}

		inputShop := &models.Shop{
			Name: "Test Shop",
		}

		expectedUser := &models.User{
			ID:    1,
			Email: "user@example.com",
		}

		signUpServiceMock := new(mocks.SignUpService)
		signUpServiceMock.EXPECT().SignUp(ctx, inputUser, inputShop).Return(expectedUser, nil)

		useCase := NewSignUpUseCase(signUpServiceMock)

		// Act
		err := useCase.Execute(ctx, inputUser, inputShop)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when sign up service fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expectedError := stdErrors.New(errors.UserAlreadyExists)

		inputUser := &models.User{
			Email:    "existing@example.com",
			Password: "password123",
		}

		inputShop := &models.Shop{
			Name: "Test Shop",
		}

		signUpServiceMock := new(mocks.SignUpService)
		signUpServiceMock.EXPECT().SignUp(ctx, inputUser, inputShop).Return(nil, expectedError)

		useCase := NewSignUpUseCase(signUpServiceMock)

		// Act
		err := useCase.Execute(ctx, inputUser, inputShop)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
