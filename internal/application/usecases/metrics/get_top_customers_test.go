package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// GetTopCustomersUseCase
// =============================================================================

func TestGetTopCustomersUseCase_Execute(t *testing.T) {
	t.Run("when no dates provided then defaults to current month", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 10}
		expected := []models.TopCustomer{{CustomerName: "John", TotalSpent: 500}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopCustomers(mock.Anything, shopID,
				mock.MatchedBy(func(from time.Time) bool {
					return from.Day() == 1 // First day of current month
				}),
				mock.AnythingOfType("time.Time"),
				10,
			).
			Return(expected, nil)

		uc := NewGetTopCustomersUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when dates provided then uses them", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
		filters := models.MetricsFilters{DateFrom: &dateFrom, DateTo: &dateTo, Limit: 10}
		expected := []models.TopCustomer{{CustomerName: "Jane"}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopCustomers(mock.Anything, shopID, dateFrom, dateTo, 10).
			Return(expected, nil)

		uc := NewGetTopCustomersUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when limit is 0 then Validated changes to 10", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 0}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopCustomers(mock.Anything, shopID, mock.Anything, mock.Anything, 10).
			Return([]models.TopCustomer{}, nil)

		uc := NewGetTopCustomersUseCase(serviceMock)

		// Act
		_, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when service errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 10}
		expectedErr := errors.New("service failure")

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopCustomers(mock.Anything, shopID, mock.Anything, mock.Anything, 10).
			Return(nil, expectedErr)

		uc := NewGetTopCustomersUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
