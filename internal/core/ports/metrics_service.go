package ports

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// MetricsService contains business logic for shop metrics calculations.
type MetricsService interface {
	// GetDashboard returns the complete dashboard metrics for a shop.
	// Executes multiple queries in parallel for performance.
	// now is the current time in the shop's timezone.
	GetDashboard(ctx context.Context, shopID int, now time.Time, tz string) (*models.DashboardMetrics, error)

	// GetRevenueTrend returns a time series of daily revenue for charting.
	GetRevenueTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.RevenueTrendPoint, error)

	// GetTopProducts returns top products with applied filters.
	GetTopProducts(ctx context.Context, shopID int, from, to time.Time, sortBy string, limit int) ([]models.TopProduct, error)

	// GetTopCustomers returns top customers with applied filters.
	GetTopCustomers(ctx context.Context, shopID int, from, to time.Time, limit int) ([]models.TopCustomer, error)

	// GetShippingSummary returns aggregated shipping metrics for a date range.
	GetShippingSummary(ctx context.Context, shopID int, from, to time.Time) (models.ShippingSummary, error)

	// GetVisitsTrend returns a time series of daily visit counts for charting.
	GetVisitsTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.VisitsTrendPoint, error)
}
