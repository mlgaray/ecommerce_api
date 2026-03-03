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
// GetTopProductsUseCase
// =============================================================================

func TestGetTopProductsUseCase_Execute(t *testing.T) {
	t.Run("when no dates provided then defaults to current month", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 10, SortBy: "quantity"}
		expected := []models.TopProduct{{ProductID: 1, ProductName: "Widget"}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopProducts(mock.Anything, shopID,
				mock.MatchedBy(func(from time.Time) bool {
					return from.Day() == 1 // First day of current month
				}),
				mock.AnythingOfType("time.Time"),
				"quantity", 10,
			).
			Return(expected, nil)

		uc := NewGetTopProductsUseCase(serviceMock)

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
		filters := models.MetricsFilters{DateFrom: &dateFrom, DateTo: &dateTo, Limit: 10, SortBy: "quantity"}
		expected := []models.TopProduct{{ProductID: 1}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopProducts(mock.Anything, shopID, dateFrom, dateTo, "quantity", 10).
			Return(expected, nil)

		uc := NewGetTopProductsUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when sortBy is revenue then passes revenue", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 10, SortBy: "revenue"}
		expected := []models.TopProduct{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopProducts(mock.Anything, shopID, mock.Anything, mock.Anything, "revenue", 10).
			Return(expected, nil)

		uc := NewGetTopProductsUseCase(serviceMock)

		// Act
		_, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when sortBy is empty then Validated changes to quantity", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 10, SortBy: ""}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopProducts(mock.Anything, shopID, mock.Anything, mock.Anything, "quantity", 10).
			Return([]models.TopProduct{}, nil)

		uc := NewGetTopProductsUseCase(serviceMock)

		// Act
		_, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when limit is 0 then Validated changes to 10", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{Limit: 0}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetTopProducts(mock.Anything, shopID, mock.Anything, mock.Anything, "quantity", 10).
			Return([]models.TopProduct{}, nil)

		uc := NewGetTopProductsUseCase(serviceMock)

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
			GetTopProducts(mock.Anything, shopID, mock.Anything, mock.Anything, "quantity", 10).
			Return(nil, expectedErr)

		uc := NewGetTopProductsUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
