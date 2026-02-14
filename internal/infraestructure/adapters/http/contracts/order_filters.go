package contracts

import (
	"strconv"
	"strings"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// OrderFiltersRequest represents HTTP query parameters for order search/filtering.
// This validates HTTP-specific concerns (format, types), NOT business rules.
// Note: ShopID is NOT included here - it's a context parameter passed separately
// from the URL path (e.g., /shops/{shop_id}/orders).
type OrderFiltersRequest struct {
	Search    *string // Optional search term (order_number or customer_name)
	Status    *string // Optional status filter
	DateFrom  *string // Optional start date (ISO format: "2006-01-02")
	DateTo    *string // Optional end date (ISO format: "2006-01-02")
	Timezone  *string // Optional IANA timezone (e.g., "America/Buenos_Aires") for date interpretation
	SortBy    string  // Sort field: "created_at", "total", "order_number"
	SortOrder string  // Sort order: "asc", "desc"
	Limit     int     // Items per page (default: 20)
	Cursor    string  // Pagination cursor (opaque base64 string)
}

// Validate validates HTTP-specific concerns.
// HTTP validations: format, types, required HTTP fields.
// Business validations happen in models.OrderFilters.Validated().
func (r *OrderFiltersRequest) Validate() error {
	// HTTP validation: if search is provided, it should not be empty or only whitespace
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		// Update to trimmed value
		*r.Search = trimmed
	}

	// HTTP validation: limit should not be negative
	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	// HTTP validation: date format validation (ISO: "2006-01-02")
	if r.DateFrom != nil {
		if _, err := time.Parse("2006-01-02", *r.DateFrom); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_date_from_format"}
		}
	}

	if r.DateTo != nil {
		if _, err := time.Parse("2006-01-02", *r.DateTo); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_date_to_format"}
		}
	}

	return nil
}

// ToOrderFilters converts HTTP request to domain model.
// Decodes the opaque cursor (if present) into LastID and LastSortValue.
// Parses date strings into time.Time values.
// Note: ShopID is NOT included - it's passed separately as a context parameter.
func (r *OrderFiltersRequest) ToOrderFilters() models.OrderFilters {
	filters := models.OrderFilters{
		Search:        r.Search,
		Status:        r.Status,
		Timezone:      r.Timezone,
		SortBy:        r.SortBy,
		SortOrder:     r.SortOrder,
		Limit:         r.Limit,
		LastID:        nil, // Will be populated from cursor
		LastSortValue: nil, // Will be populated from cursor
	}

	// Parse date_from to time.Time (start of day in UTC)
	// The use case will re-interpret these dates in the shop's timezone
	if r.DateFrom != nil {
		if dateFrom, err := time.Parse("2006-01-02", *r.DateFrom); err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	// Parse date_to to time.Time (end of day in UTC)
	// The use case will re-interpret these dates in the shop's timezone
	if r.DateTo != nil {
		if dateTo, err := time.Parse("2006-01-02", *r.DateTo); err == nil {
			endOfDay := dateTo.Add(24*time.Hour - time.Microsecond)
			filters.DateTo = &endOfDay
		}
	}

	// Decode cursor if present
	// HTTP layer responsibility: convert opaque cursor string to primitive values
	if r.Cursor != "" {
		cursorData, err := services.DecodeCursor(r.Cursor)
		if err != nil {
			// If cursor is invalid, treat as first page (LastID = nil)
			// This is graceful degradation - don't fail the request
			return filters
		}

		// Populate decoded cursor values
		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}

// NewOrderFiltersRequest parses query parameters from URL into OrderFiltersRequest.
// Note: ShopID is NOT included here - it's a context parameter parsed separately from URL path.
func NewOrderFiltersRequest(queryParams map[string][]string) (*OrderFiltersRequest, error) {
	request := &OrderFiltersRequest{}

	// Parse search
	if search := getQueryParam(queryParams, "search"); search != "" {
		request.Search = &search
	}

	// Parse status
	if status := getQueryParam(queryParams, "status"); status != "" {
		request.Status = &status
	}

	// Parse date_from
	if dateFrom := getQueryParam(queryParams, "date_from"); dateFrom != "" {
		request.DateFrom = &dateFrom
	}

	// Parse date_to
	if dateTo := getQueryParam(queryParams, "date_to"); dateTo != "" {
		request.DateTo = &dateTo
	}

	// Parse timezone (IANA identifier sent by frontend, e.g., "America/Buenos_Aires")
	if tz := getQueryParam(queryParams, "tz"); tz != "" {
		request.Timezone = &tz
	}

	// Parse sorting and pagination
	request.SortBy = getQueryParam(queryParams, "sort")
	request.SortOrder = getQueryParam(queryParams, "order")

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		request.Limit = limit
	}

	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		request.Cursor = cursorStr
	}

	return request, nil
}
