package metrics

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetTopProductsUseCase orchestrates the top products retrieval.
type GetTopProductsUseCase struct {
	metricsService ports.MetricsService
}

func NewGetTopProductsUseCase(
	metricsService ports.MetricsService,
) ports.GetTopProductsUseCase {
	return &GetTopProductsUseCase{
		metricsService: metricsService,
	}
}

// Execute validates filters, parses timezone, and delegates to the metrics service.
func (uc *GetTopProductsUseCase) Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.TopProduct, error) {
	validated := filters.Validated()

	loc, _ := parseTimezone(validated.Timezone)
	now := time.Now().In(loc)

	// Default: current month
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	to := now

	if validated.DateFrom != nil {
		from = *validated.DateFrom
	}
	if validated.DateTo != nil {
		to = *validated.DateTo
	}

	return uc.metricsService.GetTopProducts(ctx, shopID, from, to, validated.SortBy, validated.Limit)
}
