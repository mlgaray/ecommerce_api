package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetDashboardMetricsUseCase orchestrates the dashboard metrics retrieval.
type GetDashboardMetricsUseCase interface {
	Execute(ctx context.Context, shopID int, tz *string) (*models.DashboardMetrics, error)
}

// GetRevenueTrendUseCase orchestrates the revenue trend retrieval.
type GetRevenueTrendUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.RevenueTrendPoint, error)
}

// GetTopProductsUseCase orchestrates the top products retrieval.
type GetTopProductsUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.TopProduct, error)
}

// GetTopCustomersUseCase orchestrates the top customers retrieval.
type GetTopCustomersUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.TopCustomer, error)
}

// GetShippingSummaryUseCase orchestrates the shipping summary retrieval.
type GetShippingSummaryUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.MetricsFilters) (models.ShippingSummary, error)
}

// GetVisitsTrendUseCase orchestrates the visits trend retrieval.
type GetVisitsTrendUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.VisitsTrendPoint, error)
}
