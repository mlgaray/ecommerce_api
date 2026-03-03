package metrics

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetDashboardUseCase orchestrates the dashboard metrics retrieval.
// Parses the timezone sent by the frontend and delegates to the metrics service.
type GetDashboardUseCase struct {
	metricsService ports.MetricsService
}

func NewGetDashboardUseCase(
	metricsService ports.MetricsService,
) ports.GetDashboardMetricsUseCase {
	return &GetDashboardUseCase{
		metricsService: metricsService,
	}
}

// Execute parses the frontend timezone and delegates to the metrics service.
func (uc *GetDashboardUseCase) Execute(ctx context.Context, shopID int, tz *string) (*models.DashboardMetrics, error) {
	loc, tzName := parseTimezone(tz)
	now := time.Now().In(loc)

	return uc.metricsService.GetDashboard(ctx, shopID, now, tzName)
}
