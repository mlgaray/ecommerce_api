package contracts

import (
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// CategoryFiltersRequest represents HTTP query parameters for category search/filtering.
// This validates HTTP-specific concerns (format, types), NOT business rules.
// Note: ShopID comes from URL path and is passed separately to the use case.
type CategoryFiltersRequest struct {
	Search    *string // Optional search term (searches by name)
	SortBy    string  // Sort field: "name", "created_at"
	SortOrder string  // Sort order: "asc", "desc"
	Limit     int     // Items per page (default: 20)
	Cursor    string  // Pagination cursor (opaque base64 string)
}

// NewCategoryFiltersRequest creates a CategoryFiltersRequest from URL query parameters.
func NewCategoryFiltersRequest(queryParams map[string][]string) (*CategoryFiltersRequest, error) {
	request := &CategoryFiltersRequest{}

	// Parse search param
	if search := getQueryParam(queryParams, "search"); search != "" {
		request.Search = &search
	}

	// Parse sorting params
	request.SortBy = getQueryParam(queryParams, "sort")
	request.SortOrder = getQueryParam(queryParams, "order")

	// Parse limit
	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		request.Limit = limit
	}

	// Parse cursor
	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		request.Cursor = cursorStr
	}

	return request, nil
}

// Validate validates HTTP-specific concerns.
// HTTP validations: format, types.
// Business validations happen in models.CategoryFilters.Validate().
func (r *CategoryFiltersRequest) Validate() error {
	// HTTP validation: if search is provided, it should not be empty or only whitespace
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		*r.Search = trimmed
	}

	// HTTP validation: limit should be a valid integer if provided
	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	return nil
}

// ToCategoryFilters converts HTTP request to domain model.
// Decodes the opaque cursor (if present) into LastID and LastSortValue.
func (r *CategoryFiltersRequest) ToCategoryFilters() models.CategoryFilters {
	filters := models.CategoryFilters{
		Search:        r.Search,
		SortBy:        r.SortBy,
		SortOrder:     r.SortOrder,
		Limit:         r.Limit,
		LastID:        nil,
		LastSortValue: nil,
	}

	// Decode cursor if present
	if r.Cursor != "" {
		cursorData, err := services.DecodeCursor(r.Cursor)
		if err != nil {
			// If cursor is invalid, treat as first page (graceful degradation)
			return filters
		}

		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}
