package metrics

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetVisitsTrendUseCase orchestrates the visits trend retrieval.
type GetVisitsTrendUseCase struct {
	metricsService ports.MetricsService
}

func NewGetVisitsTrendUseCase(
	metricsService ports.MetricsService,
) ports.GetVisitsTrendUseCase {
	return &GetVisitsTrendUseCase{
		metricsService: metricsService,
	}
}

// Execute validates filters, parses timezone, and delegates to the metrics service.
// Defaults to last 30 days if no dates provided.
func (uc *GetVisitsTrendUseCase) Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.VisitsTrendPoint, error) {
	validated := filters.Validated()

	loc, tzName := parseTimezone(validated.Timezone)
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

	return uc.metricsService.GetVisitsTrend(ctx, shopID, from, to, tzName)
}
