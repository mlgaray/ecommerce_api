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
// GetShippingSummaryUseCase
// =============================================================================

func TestGetShippingSummaryUseCase_Execute(t *testing.T) {
	t.Run("when no dates and no tz then defaults to last 30 days in UTC", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		expected := models.ShippingSummary{TotalShippingRevenue: 5000, OrderCount: 100, AverageShippingCost: 50}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetShippingSummary(
				mock.Anything,
				shopID,
				mock.MatchedBy(func(from time.Time) bool {
					// Should be approximately 30 days ago
					diff := time.Since(from)
					return diff >= 29*24*time.Hour && diff <= 31*24*time.Hour
				}),
				mock.MatchedBy(func(to time.Time) bool {
					// Should be approximately now
					return time.Since(to) < 5*time.Second
				}),
			).
			Return(expected, nil)

		uc := NewGetShippingSummaryUseCase(serviceMock)

		result, err := uc.Execute(ctx, shopID, models.MetricsFilters{})

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when DateFrom provided then uses it", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		customFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expected := models.ShippingSummary{OrderCount: 50}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetShippingSummary(
				mock.Anything,
				shopID,
				mock.MatchedBy(func(from time.Time) bool {
					return from.Equal(customFrom)
				}),
				mock.AnythingOfType("time.Time"),
			).
			Return(expected, nil)

		uc := NewGetShippingSummaryUseCase(serviceMock)

		result, err := uc.Execute(ctx, shopID, models.MetricsFilters{DateFrom: &customFrom})

		assert.NoError(t, err)
		assert.Equal(t, 50, result.OrderCount)
	})

	t.Run("when DateTo provided then uses it", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		customTo := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)
		expected := models.ShippingSummary{OrderCount: 75}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetShippingSummary(
				mock.Anything,
				shopID,
				mock.AnythingOfType("time.Time"),
				mock.MatchedBy(func(to time.Time) bool {
					return to.Equal(customTo)
				}),
			).
			Return(expected, nil)

		uc := NewGetShippingSummaryUseCase(serviceMock)

		result, err := uc.Execute(ctx, shopID, models.MetricsFilters{DateTo: &customTo})

		assert.NoError(t, err)
		assert.Equal(t, 75, result.OrderCount)
	})

	t.Run("when invalid timezone then falls back to UTC", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		invalidTz := "Invalid/Zone"
		expected := models.ShippingSummary{OrderCount: 10}

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetShippingSummary(mock.Anything, shopID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(expected, nil)

		uc := NewGetShippingSummaryUseCase(serviceMock)

		result, err := uc.Execute(ctx, shopID, models.MetricsFilters{Timezone: &invalidTz})

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service errors then propagates error", func(t *testing.T) {
		ctx := context.Background()
		shopID := 1
		expectedErr := errors.New("service failure")

		serviceMock := mocks.NewMetricsService(t)
		serviceMock.EXPECT().
			GetShippingSummary(mock.Anything, shopID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(models.ShippingSummary{}, expectedErr)

		uc := NewGetShippingSummaryUseCase(serviceMock)

		_, err := uc.Execute(ctx, shopID, models.MetricsFilters{})

		assert.Error(t, err)
		assert.Equal(t, "service failure", err.Error())
	})
}
