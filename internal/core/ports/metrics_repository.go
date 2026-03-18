package ports

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// MetricsRepository handles aggregated read queries for shop metrics.
type MetricsRepository interface {
	// GetRevenueSummary returns revenue aggregation (total, count, AOV) for completed orders in a date range.
	GetRevenueSummary(ctx context.Context, shopID int, from, to time.Time) (models.RevenueSummary, error)

	// GetRevenueSummaryBatch returns revenue summaries for multiple date ranges in a single query.
	// Uses conditional aggregation (FILTER) for efficient single-pass computation.
	GetRevenueSummaryBatch(ctx context.Context, shopID int, periods []models.RevenuePeriod) ([]models.RevenueSummary, error)

	// GetOrderStatusDistribution returns order counts grouped by status within a date range.
	GetOrderStatusDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.OrderStatusDistribution, error)

	// GetAvgCompletionTime returns the average time in hours to complete an order within a date range.
	// Returns nil if no completed orders exist.
	GetAvgCompletionTime(ctx context.Context, shopID int, from, to time.Time) (*float64, error)

	// GetTopProducts returns top products ranked by quantity sold or revenue within a date range.
	// Only considers completed orders. sortBy must be "quantity" or "revenue".
	GetTopProducts(ctx context.Context, shopID int, from, to time.Time, sortBy string, limit int) ([]models.TopProduct, error)

	// GetCustomerOverview returns customer aggregation stats (total unique, new, returning) within a date range.
	GetCustomerOverview(ctx context.Context, shopID int, from, to time.Time) (models.CustomerOverview, error)

	// GetTopCustomers returns top customers ranked by total spend within a date range.
	// Only considers completed orders.
	GetTopCustomers(ctx context.Context, shopID int, from, to time.Time, limit int) ([]models.TopCustomer, error)

	// GetPaymentMethodDistribution returns order counts grouped by payment method within a date range.
	GetPaymentMethodDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.MethodDistribution, error)

	// GetDeliveryMethodDistribution returns order counts grouped by delivery method within a date range.
	GetDeliveryMethodDistribution(ctx context.Context, shopID int, from, to time.Time) ([]models.MethodDistribution, error)

	// GetRevenueTrend returns daily revenue data points within a date range.
	// Only considers completed orders. tz is the IANA timezone for date_trunc.
	GetRevenueTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.RevenueTrendPoint, error)

	// GetShippingSummary returns aggregated shipping metrics for completed orders with shipping in a date range.
	GetShippingSummary(ctx context.Context, shopID int, from, to time.Time) (models.ShippingSummary, error)

	// GetVisitsSummaryBatch returns visit counts for multiple date ranges in a single query.
	GetVisitsSummaryBatch(ctx context.Context, shopID int, periods []models.RevenuePeriod) ([]int, error)

	// GetVisitsTrend returns daily visit counts within a date range.
	GetVisitsTrend(ctx context.Context, shopID int, from, to time.Time, tz string) ([]models.VisitsTrendPoint, error)
}
