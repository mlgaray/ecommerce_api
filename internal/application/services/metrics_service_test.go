package services

import (
	"context"
	"errors"
	"testing"
	"time"

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
// Period Helpers - todayBounds
// =============================================================================

func TestTodayBounds(t *testing.T) {
	t.Run("when midday then returns midnight-to-midnight", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

		// Act
		from, to := todayBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC), to)
	})

	t.Run("when midnight exact then returns same day boundaries", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

		// Act
		from, to := todayBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC), to)
	})

	t.Run("when non-UTC timezone then preserves location", func(t *testing.T) {
		// Arrange
		loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
		now := time.Date(2024, 6, 15, 14, 30, 0, 0, loc)

		// Act
		from, to := todayBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 15, 0, 0, 0, 0, loc), from)
		assert.Equal(t, time.Date(2024, 6, 16, 0, 0, 0, 0, loc), to)
		assert.Equal(t, loc, from.Location())
	})
}

// =============================================================================
// Period Helpers - previousDayBounds
// =============================================================================

func TestPreviousDayBounds(t *testing.T) {
	t.Run("when normal day then returns yesterday boundaries", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

		// Act
		from, to := previousDayBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), to)
	})

	t.Run("when first of month then crosses to previous month", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)

		// Act
		from, to := previousDayBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), from) // 2024 is leap year
		assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), to)
	})
}

// =============================================================================
// Period Helpers - thisWeekBounds
// =============================================================================

func TestThisWeekBounds(t *testing.T) {
	t.Run("when wednesday then monday is 2 days back", func(t *testing.T) {
		// Arrange — 2024-06-12 is a Wednesday
		now := time.Date(2024, 6, 12, 15, 0, 0, 0, time.UTC)

		// Act
		from, to := thisWeekBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), from) // Monday
		assert.Equal(t, now, to)
	})

	t.Run("when monday then from equals today midnight", func(t *testing.T) {
		// Arrange — 2024-06-10 is a Monday
		now := time.Date(2024, 6, 10, 12, 0, 0, 0, time.UTC)

		// Act
		from, to := thisWeekBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, now, to)
	})

	t.Run("when sunday then wraps weekday 0 to 7", func(t *testing.T) {
		// Arrange — 2024-06-16 is a Sunday
		now := time.Date(2024, 6, 16, 18, 0, 0, 0, time.UTC)

		// Act
		from, to := thisWeekBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), from) // Previous Monday
		assert.Equal(t, now, to)
	})
}

// =============================================================================
// Period Helpers - previousWeekBounds
// =============================================================================

func TestPreviousWeekBounds(t *testing.T) {
	t.Run("when wednesday then returns previous monday to this monday", func(t *testing.T) {
		// Arrange — 2024-06-12 is a Wednesday
		now := time.Date(2024, 6, 12, 15, 0, 0, 0, time.UTC)

		// Act
		from, to := previousWeekBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC), from) // Previous Monday
		assert.Equal(t, time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), to)  // This Monday
	})

	t.Run("when sunday then wraps correctly", func(t *testing.T) {
		// Arrange — 2024-06-16 is a Sunday
		now := time.Date(2024, 6, 16, 18, 0, 0, 0, time.UTC)

		// Act
		from, to := previousWeekBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), to)
	})
}

// =============================================================================
// Period Helpers - thisMonthBounds
// =============================================================================

func TestThisMonthBounds(t *testing.T) {
	t.Run("when mid month then from is first day", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)

		// Act
		from, to := thisMonthBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, now, to)
	})

	t.Run("when first day of month then from equals today midnight", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

		// Act
		from, to := thisMonthBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, now, to)
	})
}

// =============================================================================
// Period Helpers - previousMonthBounds
// =============================================================================

func TestPreviousMonthBounds(t *testing.T) {
	t.Run("when march then returns february boundaries", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC)

		// Act
		from, to := previousMonthBounds(now)

		// Assert
		assert.Equal(t, time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), to)
	})

	t.Run("when january then crosses to previous year december", func(t *testing.T) {
		// Arrange
		now := time.Date(2024, 1, 10, 14, 0, 0, 0, time.UTC)

		// Act
		from, to := previousMonthBounds(now)

		// Assert
		assert.Equal(t, time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), from)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), to)
	})
}

// =============================================================================
// Business Helpers - percentChange
// =============================================================================

func TestPercentChange(t *testing.T) {
	t.Run("when previous is zero and current is positive then returns 100", func(t *testing.T) {
		assert.Equal(t, 100.0, percentChange(50, 0))
	})

	t.Run("when both are zero then returns 0", func(t *testing.T) {
		assert.Equal(t, 0.0, percentChange(0, 0))
	})

	t.Run("when increase 50 percent then returns 50", func(t *testing.T) {
		assert.Equal(t, 50.0, percentChange(150, 100))
	})

	t.Run("when decrease 50 percent then returns negative 50", func(t *testing.T) {
		assert.Equal(t, -50.0, percentChange(50, 100))
	})

	t.Run("when values are equal then returns 0", func(t *testing.T) {
		assert.Equal(t, 0.0, percentChange(100, 100))
	})
}

// =============================================================================
// Business Helpers - buildComparison
// =============================================================================

func TestBuildComparison(t *testing.T) {
	t.Run("when current greater than previous then positive changes", func(t *testing.T) {
		// Arrange
		current := models.RevenueSummary{TotalRevenue: 200, OrderCount: 10, AOV: 20}
		previous := models.RevenueSummary{TotalRevenue: 100, OrderCount: 5, AOV: 20}

		// Act
		result := buildComparison(current, previous)

		// Assert
		assert.Equal(t, current, result.Current)
		assert.Equal(t, previous, result.Previous)
		assert.Equal(t, 100.0, result.RevenueChange)
		assert.Equal(t, 100.0, result.OrderChange)
		assert.Equal(t, 0.0, result.AOVChange)
	})

	t.Run("when previous is zero then returns 100 percent changes", func(t *testing.T) {
		// Arrange
		current := models.RevenueSummary{TotalRevenue: 100, OrderCount: 5, AOV: 20}
		previous := models.RevenueSummary{}

		// Act
		result := buildComparison(current, previous)

		// Assert
		assert.Equal(t, 100.0, result.RevenueChange)
		assert.Equal(t, 100.0, result.OrderChange)
		assert.Equal(t, 100.0, result.AOVChange)
	})
}

// =============================================================================
// Business Helpers - buildOrdersMetrics
// =============================================================================

func TestBuildOrdersMetrics(t *testing.T) {
	t.Run("when statusDist is nil then returns empty slice", func(t *testing.T) {
		// Act
		result := buildOrdersMetrics(nil, nil)

		// Assert
		assert.NotNil(t, result.StatusDistribution)
		assert.Empty(t, result.StatusDistribution)
		assert.Equal(t, 0, result.TotalOrders)
		assert.Equal(t, 0.0, result.CancellationRate)
		assert.Nil(t, result.AvgCompletionTime)
	})

	t.Run("when no cancelled orders then cancellation rate is 0", func(t *testing.T) {
		// Arrange
		dist := []models.OrderStatusDistribution{
			{Status: "completed", Count: 8},
			{Status: "pending", Count: 2},
		}

		// Act
		result := buildOrdersMetrics(dist, nil)

		// Assert
		assert.Equal(t, 10, result.TotalOrders)
		assert.Equal(t, 0.0, result.CancellationRate)
	})

	t.Run("when some cancelled orders then calculates rate correctly", func(t *testing.T) {
		// Arrange
		dist := []models.OrderStatusDistribution{
			{Status: "completed", Count: 7},
			{Status: "cancelled", Count: 3},
		}

		// Act
		result := buildOrdersMetrics(dist, nil)

		// Assert
		assert.Equal(t, 10, result.TotalOrders)
		assert.Equal(t, 30.0, result.CancellationRate)
	})

	t.Run("when all cancelled then cancellation rate is 100", func(t *testing.T) {
		// Arrange
		dist := []models.OrderStatusDistribution{
			{Status: "cancelled", Count: 5},
		}

		// Act
		result := buildOrdersMetrics(dist, nil)

		// Assert
		assert.Equal(t, 100.0, result.CancellationRate)
	})

	t.Run("when avgCompletion has value then sets it", func(t *testing.T) {
		// Arrange
		avg := 24.5

		// Act
		result := buildOrdersMetrics([]models.OrderStatusDistribution{}, &avg)

		// Assert
		assert.NotNil(t, result.AvgCompletionTime)
		assert.Equal(t, 24.5, *result.AvgCompletionTime)
	})
}

// =============================================================================
// Business Helpers - calculatePercentages
// =============================================================================

func TestCalculatePercentages(t *testing.T) {
	t.Run("when empty slice then returns same slice", func(t *testing.T) {
		// Act
		result := calculatePercentages([]models.MethodDistribution{})

		// Assert
		assert.Empty(t, result)
	})

	t.Run("when nil slice then returns nil", func(t *testing.T) {
		// Act
		result := calculatePercentages(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when total is zero then returns without modifying percentages", func(t *testing.T) {
		// Arrange
		dist := []models.MethodDistribution{
			{Name: "Cash", OrderCount: 0},
		}

		// Act
		result := calculatePercentages(dist)

		// Assert
		assert.Equal(t, 0.0, result[0].Percentage)
	})

	t.Run("when single item then percentage is 100", func(t *testing.T) {
		// Arrange
		dist := []models.MethodDistribution{
			{Name: "Cash", OrderCount: 10},
		}

		// Act
		result := calculatePercentages(dist)

		// Assert
		assert.Equal(t, 100.0, result[0].Percentage)
	})

	t.Run("when multiple items then percentages sum correctly", func(t *testing.T) {
		// Arrange
		dist := []models.MethodDistribution{
			{Name: "Cash", OrderCount: 3},
			{Name: "Card", OrderCount: 7},
		}

		// Act
		result := calculatePercentages(dist)

		// Assert
		assert.Equal(t, 30.0, result[0].Percentage)
		assert.Equal(t, 70.0, result[1].Percentage)
	})
}

// =============================================================================
// Service Methods - GetDashboard
// =============================================================================

func newMockRevenueSummaries() []models.RevenueSummary {
	return []models.RevenueSummary{
		{TotalRevenue: 1000, OrderCount: 10, AOV: 100},   // today
		{TotalRevenue: 800, OrderCount: 8, AOV: 100},     // yesterday
		{TotalRevenue: 5000, OrderCount: 50, AOV: 100},   // this week
		{TotalRevenue: 4000, OrderCount: 40, AOV: 100},   // prev week
		{TotalRevenue: 20000, OrderCount: 200, AOV: 100}, // this month
		{TotalRevenue: 18000, OrderCount: 180, AOV: 100}, // prev month
	}
}

func TestMetricsService_GetDashboard(t *testing.T) {
	t.Run("when all repo calls succeed then returns complete dashboard", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
		avgComp := 24.0

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueSummaryBatch(ctx, shopID, mock.AnythingOfType("[]models.RevenuePeriod")).Return(newMockRevenueSummaries(), nil)
		repoMock.EXPECT().GetOrderStatusDistribution(ctx, shopID, mock.Anything, mock.Anything).Return([]models.OrderStatusDistribution{{Status: "completed", Count: 10}}, nil)
		repoMock.EXPECT().GetAvgCompletionTime(ctx, shopID, mock.Anything, mock.Anything).Return(&avgComp, nil)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, mock.Anything, mock.Anything, "quantity", 5).Return([]models.TopProduct{{ProductID: 1}}, nil)
		repoMock.EXPECT().GetPaymentMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return([]models.MethodDistribution{{Name: "Cash", OrderCount: 10}}, nil)
		repoMock.EXPECT().GetDeliveryMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return([]models.MethodDistribution{{Name: "Delivery", OrderCount: 5}}, nil)
		repoMock.EXPECT().GetCustomerOverview(ctx, shopID, mock.Anything, mock.Anything).Return(models.CustomerOverview{TotalUnique: 50}, nil)
		repoMock.EXPECT().GetVisitsSummaryBatch(ctx, shopID, mock.AnythingOfType("[]models.RevenuePeriod")).Return([]int{10, 8, 50, 40, 200, 180}, nil)

		service := NewMetricsService(repoMock)

		// Act
		dashboard, err := service.GetDashboard(ctx, shopID, now, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		assert.Equal(t, 1000.0, dashboard.Revenue.Today.Current.TotalRevenue)
		assert.Equal(t, 10, dashboard.Orders.TotalOrders)
		assert.Len(t, dashboard.TopProducts, 1)
		assert.Len(t, dashboard.PaymentDistribution, 1)
		assert.Len(t, dashboard.DeliveryDistribution, 1)
		assert.Equal(t, 50, dashboard.Customers.TotalUnique)
		assert.Equal(t, 10, dashboard.Visits.Today.Current)
		assert.Equal(t, 8, dashboard.Visits.Today.Previous)
	})

	t.Run("when revenueBatch errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
		expectedErr := errors.New("db connection failed")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueSummaryBatch(ctx, shopID, mock.Anything).Return(nil, expectedErr)
		repoMock.EXPECT().GetOrderStatusDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetAvgCompletionTime(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetTopProducts(ctx, shopID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetPaymentMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetDeliveryMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetCustomerOverview(ctx, shopID, mock.Anything, mock.Anything).Return(models.CustomerOverview{}, nil).Maybe()
		repoMock.EXPECT().GetVisitsSummaryBatch(ctx, shopID, mock.Anything).Return(nil, nil).Maybe()

		service := NewMetricsService(repoMock)

		// Act
		dashboard, err := service.GetDashboard(ctx, shopID, now, "UTC")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, dashboard)
		assert.Equal(t, "db connection failed", err.Error())
	})

	t.Run("when statusDistribution errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
		expectedErr := errors.New("status query failed")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueSummaryBatch(ctx, shopID, mock.Anything).Return(newMockRevenueSummaries(), nil).Maybe()
		repoMock.EXPECT().GetOrderStatusDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, expectedErr)
		repoMock.EXPECT().GetAvgCompletionTime(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetTopProducts(ctx, shopID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetPaymentMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetDeliveryMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
		repoMock.EXPECT().GetCustomerOverview(ctx, shopID, mock.Anything, mock.Anything).Return(models.CustomerOverview{}, nil).Maybe()
		repoMock.EXPECT().GetVisitsSummaryBatch(ctx, shopID, mock.Anything).Return(nil, nil).Maybe()

		service := NewMetricsService(repoMock)

		// Act
		dashboard, err := service.GetDashboard(ctx, shopID, now, "UTC")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, dashboard)
	})

	t.Run("when topProducts is nil then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueSummaryBatch(ctx, shopID, mock.Anything).Return(newMockRevenueSummaries(), nil)
		repoMock.EXPECT().GetOrderStatusDistribution(ctx, shopID, mock.Anything, mock.Anything).Return([]models.OrderStatusDistribution{}, nil)
		repoMock.EXPECT().GetAvgCompletionTime(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, mock.Anything, mock.Anything, "quantity", 5).Return(nil, nil)
		repoMock.EXPECT().GetPaymentMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetDeliveryMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetCustomerOverview(ctx, shopID, mock.Anything, mock.Anything).Return(models.CustomerOverview{}, nil)
		repoMock.EXPECT().GetVisitsSummaryBatch(ctx, shopID, mock.AnythingOfType("[]models.RevenuePeriod")).Return([]int{0, 0, 0, 0, 0, 0}, nil)

		service := NewMetricsService(repoMock)

		// Act
		dashboard, err := service.GetDashboard(ctx, shopID, now, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, dashboard.TopProducts)
		assert.Empty(t, dashboard.TopProducts)
		assert.NotNil(t, dashboard.PaymentDistribution)
		assert.Empty(t, dashboard.PaymentDistribution)
		assert.NotNil(t, dashboard.DeliveryDistribution)
		assert.Empty(t, dashboard.DeliveryDistribution)
	})
}

// =============================================================================
// Service Methods - GetRevenueTrend
// =============================================================================

func TestMetricsService_GetRevenueTrend(t *testing.T) {
	t.Run("when repo returns data then returns data", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expected := []models.RevenueTrendPoint{{Date: "2024-06-01", Revenue: 100, OrderCount: 5}}

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueTrend(ctx, shopID, from, to, "UTC").Return(expected, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetRevenueTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when repo returns nil then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueTrend(ctx, shopID, from, to, "UTC").Return(nil, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetRevenueTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("when repo errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expectedErr := errors.New("db error")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueTrend(ctx, shopID, from, to, "UTC").Return(nil, expectedErr)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetRevenueTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Service Methods - GetTopProducts
// =============================================================================

func TestMetricsService_GetTopProducts(t *testing.T) {
	t.Run("when repo returns data then returns data", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expected := []models.TopProduct{{ProductID: 1, ProductName: "Widget", QuantitySold: 100, Revenue: 999.99}}

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, from, to, "quantity", 10).Return(expected, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopProducts(ctx, shopID, from, to, "quantity", 10)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when repo returns nil then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, from, to, "quantity", 10).Return(nil, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopProducts(ctx, shopID, from, to, "quantity", 10)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("when repo errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expectedErr := errors.New("db error")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, from, to, "quantity", 10).Return(nil, expectedErr)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopProducts(ctx, shopID, from, to, "quantity", 10)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Service Methods - GetVisitsTrend
// =============================================================================

func TestMetricsService_GetVisitsTrend(t *testing.T) {
	t.Run("when repo returns data then returns data", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expected := []models.VisitsTrendPoint{{Date: "2024-06-01", VisitCount: 42}}

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetVisitsTrend(ctx, shopID, from, to, "UTC").Return(expected, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetVisitsTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when repo returns nil then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetVisitsTrend(ctx, shopID, from, to, "UTC").Return(nil, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetVisitsTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("when repo errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expectedErr := errors.New("db error")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetVisitsTrend(ctx, shopID, from, to, "UTC").Return(nil, expectedErr)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetVisitsTrend(ctx, shopID, from, to, "UTC")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// =============================================================================
// Business Helpers - buildVisitsComparison
// =============================================================================

func TestBuildVisitsComparison(t *testing.T) {
	t.Run("when current greater than previous then positive change", func(t *testing.T) {
		// Act
		result := buildVisitsComparison(200, 100)

		// Assert
		assert.Equal(t, 200, result.Current)
		assert.Equal(t, 100, result.Previous)
		assert.Equal(t, 100.0, result.Change)
	})

	t.Run("when previous is zero and current positive then change is 100", func(t *testing.T) {
		// Act
		result := buildVisitsComparison(50, 0)

		// Assert
		assert.Equal(t, 50, result.Current)
		assert.Equal(t, 0, result.Previous)
		assert.Equal(t, 100.0, result.Change)
	})

	t.Run("when both are zero then change is 0", func(t *testing.T) {
		// Act
		result := buildVisitsComparison(0, 0)

		// Assert
		assert.Equal(t, 0, result.Current)
		assert.Equal(t, 0, result.Previous)
		assert.Equal(t, 0.0, result.Change)
	})
}

// =============================================================================
// Service Methods - GetDashboard (visits batch guard)
// =============================================================================

func TestMetricsService_GetDashboard_VisitsBatchGuard(t *testing.T) {
	t.Run("when visits batch returns fewer elements than periods then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		now := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetRevenueSummaryBatch(ctx, shopID, mock.Anything).Return(newMockRevenueSummaries(), nil)
		repoMock.EXPECT().GetOrderStatusDistribution(ctx, shopID, mock.Anything, mock.Anything).Return([]models.OrderStatusDistribution{}, nil)
		repoMock.EXPECT().GetAvgCompletionTime(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetTopProducts(ctx, shopID, mock.Anything, mock.Anything, "quantity", 5).Return(nil, nil)
		repoMock.EXPECT().GetPaymentMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetDeliveryMethodDistribution(ctx, shopID, mock.Anything, mock.Anything).Return(nil, nil)
		repoMock.EXPECT().GetCustomerOverview(ctx, shopID, mock.Anything, mock.Anything).Return(models.CustomerOverview{}, nil)
		// Return only 3 counts instead of the expected 6 (one per period)
		repoMock.EXPECT().GetVisitsSummaryBatch(ctx, shopID, mock.Anything).Return([]int{10, 8, 50}, nil)

		service := NewMetricsService(repoMock)

		// Act
		dashboard, err := service.GetDashboard(ctx, shopID, now, "UTC")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, dashboard)
		assert.Contains(t, err.Error(), "visits summary batch returned")
	})
}

// =============================================================================
// Service Methods - GetTopCustomers
// =============================================================================

func TestMetricsService_GetTopCustomers(t *testing.T) {
	t.Run("when repo returns data then returns data", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expected := []models.TopCustomer{{CustomerName: "John", OrderCount: 5, TotalSpent: 500}}

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopCustomers(ctx, shopID, from, to, 10).Return(expected, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopCustomers(ctx, shopID, from, to, 10)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when repo returns nil then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopCustomers(ctx, shopID, from, to, 10).Return(nil, nil)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopCustomers(ctx, shopID, from, to, 10)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("when repo errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		from := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		expectedErr := errors.New("db error")

		repoMock := mocks.NewMetricsRepository(t)
		repoMock.EXPECT().GetTopCustomers(ctx, shopID, from, to, 10).Return(nil, expectedErr)

		service := NewMetricsService(repoMock)

		// Act
		result, err := service.GetTopCustomers(ctx, shopID, from, to, 10)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
