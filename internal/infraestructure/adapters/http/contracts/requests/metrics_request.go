package requests

import (
	"strconv"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// MetricsFiltersRequest represents HTTP query parameters for metrics endpoints.
type MetricsFiltersRequest struct {
	DateFrom *string
	DateTo   *string
	Timezone *string
	Limit    int
	SortBy   string // "quantity" | "revenue" (only for top products)
}

// loadTimezone returns the timezone location for date parsing.
func (r *MetricsFiltersRequest) loadTimezone() *time.Location {
	if r.Timezone != nil && *r.Timezone != "" {
		if loc, err := time.LoadLocation(*r.Timezone); err == nil {
			return loc
		}
	}
	return time.UTC
}

// Validate validates HTTP-specific concerns (format, types).
func (r *MetricsFiltersRequest) Validate() error {
	loc := r.loadTimezone()

	if r.DateFrom != nil {
		if _, err := time.ParseInLocation("2006-01-02", *r.DateFrom, loc); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_date_from_format"}
		}
	}

	if r.DateTo != nil {
		if _, err := time.ParseInLocation("2006-01-02", *r.DateTo, loc); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_date_to_format"}
		}
	}

	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	return nil
}

// ToMetricsFilters converts HTTP request to domain model.
func (r *MetricsFiltersRequest) ToMetricsFilters() models.MetricsFilters {
	filters := models.MetricsFilters{
		Timezone: r.Timezone,
		Limit:    r.Limit,
		SortBy:   r.SortBy,
	}

	loc := r.loadTimezone()

	if r.DateFrom != nil {
		if dateFrom, err := time.ParseInLocation("2006-01-02", *r.DateFrom, loc); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if r.DateTo != nil {
		if dateTo, err := time.ParseInLocation("2006-01-02", *r.DateTo, loc); err == nil {
			nextDay := dateTo.AddDate(0, 0, 1)
			filters.DateTo = &nextDay
		}
	}

	return filters
}

// NewMetricsFiltersRequest parses query parameters into a MetricsFiltersRequest.
func NewMetricsFiltersRequest(queryParams map[string][]string) (*MetricsFiltersRequest, error) {
	request := &MetricsFiltersRequest{}

	if dateFrom := getQueryParam(queryParams, "date_from"); dateFrom != "" {
		request.DateFrom = &dateFrom
	}

	if dateTo := getQueryParam(queryParams, "date_to"); dateTo != "" {
		request.DateTo = &dateTo
	}

	if tz := getQueryParam(queryParams, "tz"); tz != "" {
		request.Timezone = &tz
	}

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		request.Limit = limit
	}

	if sortBy := getQueryParam(queryParams, "sort_by"); sortBy != "" {
		request.SortBy = sortBy
	}

	return request, nil
}
