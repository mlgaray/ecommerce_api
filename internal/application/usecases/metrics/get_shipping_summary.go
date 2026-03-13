package metrics

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetShippingSummaryUseCase orchestrates the shipping summary retrieval.
type GetShippingSummaryUseCase struct {
	metricsService ports.MetricsService
}

func NewGetShippingSummaryUseCase(
	metricsService ports.MetricsService,
) ports.GetShippingSummaryUseCase {
	return &GetShippingSummaryUseCase{
		metricsService: metricsService,
	}
}

// Execute validates filters, parses timezone, and delegates to the metrics service.
// Defaults to last 30 days if no dates provided.
func (uc *GetShippingSummaryUseCase) Execute(ctx context.Context, shopID int, filters models.MetricsFilters) (models.ShippingSummary, error) {
	validated := filters.Validated()

	loc, _ := parseTimezone(validated.Timezone)
	now := time.Now().In(loc)

	// Default: last 30 days
	from := now.AddDate(0, 0, -30)
	to := now

	if validated.DateFrom != nil {
		from = *validated.DateFrom
	}
	if validated.DateTo != nil {
		to = *validated.DateTo
	}

	return uc.metricsService.GetShippingSummary(ctx, shopID, from, to)
}
