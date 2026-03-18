package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// CheckSlugAvailabilityUseCase Tests
// =============================================================================

func TestCheckSlugAvailabilityUseCase_Execute(t *testing.T) {
	t.Run("when slug is available then returns true", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "new-store"

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			CheckSlugAvailability(ctx, slug).
			Return(true, nil)

		useCase := NewCheckSlugAvailabilityUseCase(serviceMock)

		// Act
		available, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.True(t, available)
	})

	t.Run("when slug is taken then returns false", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "existing-store"

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			CheckSlugAvailability(ctx, slug).
			Return(false, nil)

		useCase := NewCheckSlugAvailabilityUseCase(serviceMock)

		// Act
		available, err := useCase.Execute(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.False(t, available)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "some-store"
		expectedError := errors.New("database error")

		serviceMock := mocks.NewStoreService(t)
		serviceMock.EXPECT().
			CheckSlugAvailability(ctx, slug).
			Return(false, expectedError)

		useCase := NewCheckSlugAvailabilityUseCase(serviceMock)

		// Act
		available, err := useCase.Execute(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.False(t, available)
		assert.Equal(t, expectedError, err)
	})
}
