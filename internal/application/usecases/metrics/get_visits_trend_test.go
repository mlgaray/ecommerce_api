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
// GetVisitsTrendUseCase
// =============================================================================

func TestGetVisitsTrendUseCase_Execute(t *testing.T) {
	t.Run("when no dates provided then defaults to last 30 days", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{}
		expected := []models.VisitsTrendPoint{{Date: "2024-06-01", VisitCount: 42}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID,
				mock.MatchedBy(func(from time.Time) bool {
					// Should be approximately 30 days ago
					diff := time.Since(from)
					return diff > 29*24*time.Hour && diff < 31*24*time.Hour
				}),
				mock.AnythingOfType("time.Time"),
				"UTC",
			).
			Return(expected, nil)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when DateFrom and DateTo provided then uses them", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
		filters := models.MetricsFilters{DateFrom: &dateFrom, DateTo: &dateTo}
		expected := []models.VisitsTrendPoint{{Date: "2024-01-15", VisitCount: 10}}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID, dateFrom, dateTo, "UTC").
			Return(expected, nil)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when only DateFrom provided then uses it with default to", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		filters := models.MetricsFilters{DateFrom: &dateFrom}
		expected := []models.VisitsTrendPoint{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID, dateFrom, mock.AnythingOfType("time.Time"), "UTC").
			Return(expected, nil)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		_, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when only DateTo provided then uses default from with DateTo", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		dateTo := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		filters := models.MetricsFilters{DateTo: &dateTo}
		expected := []models.VisitsTrendPoint{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID, mock.AnythingOfType("time.Time"), dateTo, "UTC").
			Return(expected, nil)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		_, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when service errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.MetricsFilters{}
		expectedErr := errors.New("service failure")

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID, mock.Anything, mock.Anything, "UTC").
			Return(nil, expectedErr)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("when timezone provided then parses correctly", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		tz := "America/Argentina/Buenos_Aires"
		filters := models.MetricsFilters{Timezone: &tz}
		expected := []models.VisitsTrendPoint{}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetVisitsTrend(mock.Anything, shopID, mock.Anything, mock.Anything, "America/Argentina/Buenos_Aires").
			Return(expected, nil)

		uc := NewGetVisitsTrendUseCase(serviceMock)

		// Act
		result, err := uc.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
