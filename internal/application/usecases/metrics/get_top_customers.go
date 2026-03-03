package metrics

import (
	"context"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetTopCustomersUseCase orchestrates the top customers retrieval.
type GetTopCustomersUseCase struct {
	metricsService ports.MetricsService
}

func NewGetTopCustomersUseCase(
	metricsService ports.MetricsService,
) ports.GetTopCustomersUseCase {
	return &GetTopCustomersUseCase{
		metricsService: metricsService,
	}
}

// Execute validates filters, parses timezone, and delegates to the metrics service.
func (uc *GetTopCustomersUseCase) Execute(ctx context.Context, shopID int, filters models.MetricsFilters) ([]models.TopCustomer, error) {
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

	return uc.metricsService.GetTopCustomers(ctx, shopID, from, to, validated.Limit)
}
