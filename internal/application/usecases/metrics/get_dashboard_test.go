package metrics

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func init() {
	logs.Init()
}

// =============================================================================
// GetDashboardUseCase
// =============================================================================

func TestGetDashboardUseCase_Execute(t *testing.T) {
	t.Run("when tz is nil then calls service with UTC", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expected := &models.DashboardMetrics{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetDashboard(mock.Anything, shopID, mock.AnythingOfType("time.Time"), "UTC").
			Return(expected, nil)

		uc := NewGetDashboardUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when tz is valid then calls service with timezone name", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		tz := "America/New_York"
		expected := &models.DashboardMetrics{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetDashboard(mock.Anything, shopID, mock.AnythingOfType("time.Time"), "America/New_York").
			Return(expected, nil)

		uc := NewGetDashboardUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, &tz)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when tz is invalid then falls back to UTC", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		tz := "Invalid/Zone"
		expected := &models.DashboardMetrics{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetDashboard(mock.Anything, shopID, mock.AnythingOfType("time.Time"), "UTC").
			Return(expected, nil)

		uc := NewGetDashboardUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, &tz)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedErr := errors.New("service failure")

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetDashboard(mock.Anything, shopID, mock.AnythingOfType("time.Time"), "UTC").
			Return(nil, expectedErr)

		uc := NewGetDashboardUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "service failure", err.Error())
	})
}
