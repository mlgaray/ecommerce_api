package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// MetricsService contains business logic for shop metrics calculations.
type MetricsService struct {
	metricsRepository ports.MetricsRepository
}

func NewMetricsService(metricsRepository ports.MetricsRepository) *MetricsService {
	return &MetricsService{
		metricsRepository: metricsRepository,
	}
}

// GetDashboard returns the complete dashboard metrics for a shop.
// Executes multiple queries in parallel for performance.
func (s *MetricsService) GetDashboard(ctx context.Context, shopID int, now time.Time, tz string) (*models.DashboardMetrics, error) {
	// Calculate period boundaries in shop's timezone
	todayFrom, todayTo := todayBounds(now)
	yesterdayFrom, yesterdayTo := previousDayBounds(now)
	weekFrom, weekTo := thisWeekBounds(now)
	prevWeekFrom, prevWeekTo := previousWeekBounds(now)
	monthFrom, monthTo := thisMonthBounds(now)
	prevMonthFrom, prevMonthTo := previousMonthBounds(now)

	// Revenue periods: consolidated into a single batch query (6 periods → 1 DB call)
	periods := []models.RevenuePeriod{
		{From: todayFrom, To: todayTo},         // [0] today
		{From: yesterdayFrom, To: yesterdayTo}, // [1] yesterday
		{From: weekFrom, To: weekTo},           // [2] this week
		{From: prevWeekFrom, To: prevWeekTo},   // [3] previous week
		{From: monthFrom, To: monthTo},         // [4] this month
		{From: prevMonthFrom, To: prevMonthTo}, // [5] previous month
	}

	var (
		wg sync.WaitGroup

		// Revenue batch (1 query instead of 6)
		revenueSummaries []models.RevenueSummary
		revenueErr       error

		// Orders metrics (2 queries)
		statusDist    []models.OrderStatusDistribution
		avgCompletion *float64
		statusErr     error
		avgCompErr    error

		// Top products (1 query)
		topProducts []models.TopProduct
		topProdErr  error

		// Payment/delivery distribution (2 queries)
		paymentDist  []models.MethodDistribution
		deliveryDist []models.MethodDistribution
		payDistErr   error
		delDistErr   error

		// Customer overview (1 query)
		customers    models.CustomerOverview
		customersErr error

		// Visits batch (1 query)
		visitCounts []int
		visitsErr   error
	)

	// Execute all queries in parallel (7 goroutines instead of 12)
	wg.Go(func() {
		revenueSummaries, revenueErr = s.metricsRepository.GetRevenueSummaryBatch(ctx, shopID, periods)
	})
	wg.Go(func() {
		statusDist, statusErr = s.metricsRepository.GetOrderStatusDistribution(ctx, shopID, monthFrom, monthTo)
	})
	wg.Go(func() {
		avgCompletion, avgCompErr = s.metricsRepository.GetAvgCompletionTime(ctx, shopID, monthFrom, monthTo)
	})
	wg.Go(func() {
		topProducts, topProdErr = s.metricsRepository.GetTopProducts(ctx, shopID, monthFrom, monthTo, "quantity", 5)
	})
	wg.Go(func() {
		paymentDist, payDistErr = s.metricsRepository.GetPaymentMethodDistribution(ctx, shopID, monthFrom, monthTo)
	})
	wg.Go(func() {
		deliveryDist, delDistErr = s.metricsRepository.GetDeliveryMethodDistribution(ctx, shopID, monthFrom, monthTo)
	})
	wg.Go(func() {
		customers, customersErr = s.metricsRepository.GetCustomerOverview(ctx, shopID, monthFrom, monthTo)
	})
	wg.Go(func() {
		visitCounts, visitsErr = s.metricsRepository.GetVisitsSummaryBatch(ctx, shopID, periods)
	})

	wg.Wait()

	// Check for errors (return first encountered)
	for _, err := range []error{revenueErr, statusErr, avgCompErr, topProdErr, payDistErr, delDistErr, customersErr, visitsErr} {
		if err != nil {
			return nil, err
		}
	}

	// Guard against unexpected slice length from repository
	if len(visitCounts) < len(periods) {
		return nil, fmt.Errorf("visits summary batch returned %d counts, expected %d", len(visitCounts), len(periods))
	}

	// Build visits summary from batch counts (same period indices as revenue)
	visits := models.VisitsSummary{
		Today:     buildVisitsComparison(visitCounts[0], visitCounts[1]),
		ThisWeek:  buildVisitsComparison(visitCounts[2], visitCounts[3]),
		ThisMonth: buildVisitsComparison(visitCounts[4], visitCounts[5]),
	}

	// Build dashboard
	dashboard := &models.DashboardMetrics{
		Revenue: models.RevenueMetrics{
			Today:     buildComparison(revenueSummaries[0], revenueSummaries[1]),
			ThisWeek:  buildComparison(revenueSummaries[2], revenueSummaries[3]),
			ThisMonth: buildComparison(revenueSummaries[4], revenueSummaries[5]),
		},
		Orders:               buildOrdersMetrics(statusDist, avgCompletion),
		TopProducts:          topProducts,
		PaymentDistribution:  calculatePercentages(paymentDist),
		DeliveryDistribution: calculatePercentages(deliveryDist),
		Customers:            customers,
		Visits:               visits,
	}

	if dashboard.TopProducts == nil {
		dashboard.TopProducts = []models.TopProduct{}
	}
	if dashboard.PaymentDistribution == nil {
		dashboard.PaymentDistribution = []models.MethodDistribution{}
	}
	if dashboard.DeliveryDistribution == nil {
		dashboard.DeliveryDistribution = []models.MethodDistribution{}
	}

	return dashboard, nil
}

// GetRevenueTrend returns a time series of daily revenue for charting.
func (s *MetricsService) GetRevenueTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.RevenueTrendPoint, error) {
	trend, err := s.metricsRepository.GetRevenueTrend(ctx, shopID, from, to, tz)
	if err != nil {
		return nil, err
	}
	if trend == nil {
		return []models.RevenueTrendPoint{}, nil
	}
	return trend, nil
}

// GetTopProducts returns top products with applied filters.
func (s *MetricsService) GetTopProducts(ctx context.Context, shopID int, from, to time.Time, sortBy string, limit int) ([]models.TopProduct, error) {
	products, err := s.metricsRepository.GetTopProducts(ctx, shopID, from, to, sortBy, limit)
	if err != nil {
		return nil, err
	}
	if products == nil {
		return []models.TopProduct{}, nil
	}
	return products, nil
}

// GetTopCustomers returns top customers with applied filters.
func (s *MetricsService) GetTopCustomers(ctx context.Context, shopID int, from, to time.Time, limit int) ([]models.TopCustomer, error) {
	customers, err := s.metricsRepository.GetTopCustomers(ctx, shopID, from, to, limit)
	if err != nil {
		return nil, err
	}
	if customers == nil {
		return []models.TopCustomer{}, nil
	}
	return customers, nil
}

// GetShippingSummary returns aggregated shipping metrics for a date range.
func (s *MetricsService) GetShippingSummary(ctx context.Context, shopID int, from, to time.Time) (models.ShippingSummary, error) {
	return s.metricsRepository.GetShippingSummary(ctx, shopID, from, to)
}

// GetVisitsTrend returns a time series of daily visit counts for charting.
func (s *MetricsService) GetVisitsTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.VisitsTrendPoint, error) {
	trend, err := s.metricsRepository.GetVisitsTrend(ctx, shopID, from, to, tz)
	if err != nil {
		return nil, err
	}
	if trend == nil {
		return []models.VisitsTrendPoint{}, nil
	}
	return trend, nil
}

// --- Period calculation helpers (pure functions) ---

// todayBounds returns [start of today, start of tomorrow) in the given timezone.
func todayBounds(now time.Time) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()
	from := time.Date(y, m, d, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 0, 1)
	return from, to
}

// previousDayBounds returns [start of yesterday, start of today) in the given timezone.
func previousDayBounds(now time.Time) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()
	to := time.Date(y, m, d, 0, 0, 0, 0, loc)
	from := to.AddDate(0, 0, -1)
	return from, to
}

// thisWeekBounds returns [Monday 00:00, now) for the current ISO week (Monday-Sunday).
func thisWeekBounds(now time.Time) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)

	// Go's Weekday: Sunday=0, Monday=1, ..., Saturday=6
	// We want Monday=0
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := today.AddDate(0, 0, -(weekday - 1))
	return monday, now
}

// previousWeekBounds returns [previous Monday 00:00, this Monday 00:00).
func previousWeekBounds(now time.Time) (time.Time, time.Time) {
	y, m, d := now.Date()
	loc := now.Location()
	today := time.Date(y, m, d, 0, 0, 0, 0, loc)

	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := today.AddDate(0, 0, -(weekday - 1))
	prevMonday := thisMonday.AddDate(0, 0, -7)
	return prevMonday, thisMonday
}

// thisMonthBounds returns [1st of month 00:00, now) in the given timezone.
func thisMonthBounds(now time.Time) (time.Time, time.Time) {
	y, m, _ := now.Date()
	loc := now.Location()
	from := time.Date(y, m, 1, 0, 0, 0, 0, loc)
	return from, now
}

// previousMonthBounds returns [1st of previous month, 1st of current month).
func previousMonthBounds(now time.Time) (time.Time, time.Time) {
	y, m, _ := now.Date()
	loc := now.Location()
	to := time.Date(y, m, 1, 0, 0, 0, 0, loc)
	from := to.AddDate(0, -1, 0)
	return from, to
}

// --- Business logic helpers ---

// buildComparison builds a RevenuePeriodComparison from current and previous summaries.
func buildComparison(current, previous models.RevenueSummary) models.RevenuePeriodComparison {
	return models.RevenuePeriodComparison{
		Current:       current,
		Previous:      previous,
		RevenueChange: percentChange(current.TotalRevenue, previous.TotalRevenue),
		OrderChange:   percentChange(float64(current.OrderCount), float64(previous.OrderCount)),
		AOVChange:     percentChange(current.AOV, previous.AOV),
	}
}

// buildVisitsComparison builds a VisitsPeriodComparison from current and previous visit counts.
func buildVisitsComparison(current, previous int) models.VisitsPeriodComparison {
	return models.VisitsPeriodComparison{
		Current:  current,
		Previous: previous,
		Change:   percentChange(float64(current), float64(previous)),
	}
}

// percentChange calculates ((current - previous) / previous) * 100.
// Returns 0 if previous is zero (no comparison possible).
func percentChange(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return ((current - previous) / previous) * 100
}

// buildOrdersMetrics creates OrdersMetrics from status distribution and avg completion time.
func buildOrdersMetrics(statusDist []models.OrderStatusDistribution, avgCompletion *float64) models.OrdersMetrics {
	if statusDist == nil {
		statusDist = []models.OrderStatusDistribution{}
	}

	var total, cancelled int
	for _, s := range statusDist {
		total += s.Count
		if s.Status == "cancelled" {
			cancelled = s.Count
		}
	}

	var cancellationRate float64
	if total > 0 {
		cancellationRate = (float64(cancelled) / float64(total)) * 100
	}

	return models.OrdersMetrics{
		TotalOrders:        total,
		StatusDistribution: statusDist,
		CancellationRate:   cancellationRate,
		AvgCompletionTime:  avgCompletion,
	}
}

// calculatePercentages converts raw order counts to percentages.
func calculatePercentages(distribution []models.MethodDistribution) []models.MethodDistribution {
	if len(distribution) == 0 {
		return distribution
	}

	var total int
	for _, d := range distribution {
		total += d.OrderCount
	}

	if total == 0 {
		return distribution
	}

	for i := range distribution {
		distribution[i].Percentage = (float64(distribution[i].OrderCount) / float64(total)) * 100
	}

	return distribution
}
