package contracts

import (
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// ProductFiltersRequest represents HTTP query parameters for product search/filtering
// This validates HTTP-specific concerns (format, types), NOT business rules
// Note: ShopID is NOT included here - it's a context parameter passed separately
// from the URL path (e.g., /shops/{shop_id}/products)
type ProductFiltersRequest struct {
	Search        *string  // Optional search term
	CategoryID    *int     // Optional category filter
	IsActive      *bool    // Optional active status filter
	IsHighlighted *bool    // Optional highlighted filter
	IsPromotional *bool    // Optional promotional filter
	MinPrice      *float64 // Optional minimum price
	MaxPrice      *float64 // Optional maximum price
	SortBy        string   // Sort field: "price", "name", "created_at"
	SortOrder     string   // Sort order: "asc", "desc"
	Limit         int      // Items per page (default: 20)
	Cursor        string   // Pagination cursor (opaque base64 string)
}

// Validate validates HTTP-specific concerns
// HTTP validations: format, types, required HTTP fields
// Business validations happen in models.ProductFilters.Validate()
// Note: ShopID validation is NOT here - it's validated separately in the handler
func (r *ProductFiltersRequest) Validate() error {
	// HTTP validation: if search is provided, it should not be empty or only whitespace
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		// Update to trimmed value
		*r.Search = trimmed
	}

	// HTTP validation: limit should be a valid integer if provided
	// Actual business rules (min/max) are handled in domain layer
	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	// HTTP validation: cursor is an opaque string, no format validation needed here
	// Decoding happens in ToProductFilters() method - invalid cursors are treated as first page

	// HTTP validation: prices should be valid numbers if provided
	if r.MinPrice != nil && *r.MinPrice < 0 {
		return &httpErrors.BadRequestError{Message: "min_price_cannot_be_negative"}
	}

	if r.MaxPrice != nil && *r.MaxPrice < 0 {
		return &httpErrors.BadRequestError{Message: "max_price_cannot_be_negative"}
	}

	// All other business rules (price range logic, sort field validation, etc.)
	// are handled in models.ProductFilters.Validate()

	return nil
}

// ToProductFilters converts HTTP request to domain model
// Decodes the opaque cursor (if present) into LastID and LastSortValue
// Note: ShopID is NOT included - it's passed separately as a context parameter
func (r *ProductFiltersRequest) ToProductFilters() models.ProductFilters {
	filters := models.ProductFilters{
		Search:        r.Search,
		CategoryID:    r.CategoryID,
		IsActive:      r.IsActive,
		IsHighlighted: r.IsHighlighted,
		IsPromotional: r.IsPromotional,
		MinPrice:      r.MinPrice,
		MaxPrice:      r.MaxPrice,
		SortBy:        r.SortBy,
		SortOrder:     r.SortOrder,
		Limit:         r.Limit,
		LastID:        nil, // Will be populated from cursor
		LastSortValue: nil, // Will be populated from cursor
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

// NewProductFiltersRequest parses query parameters from URL into ProductFiltersRequest
// Note: ShopID is NOT included here - it's a context parameter parsed separately from URL path
func NewProductFiltersRequest(queryParams map[string][]string) (*ProductFiltersRequest, error) {
	request := &ProductFiltersRequest{}

	request.parseSearchParam(queryParams)

	if err := request.parseFilterParams(queryParams); err != nil {
		return nil, err
	}

	if err := request.parsePriceParams(queryParams); err != nil {
		return nil, err
	}

	if err := request.parsePaginationParams(queryParams); err != nil {
		return nil, err
	}

	return request, nil
}

// parseSearchParam extracts search query parameter
func (r *ProductFiltersRequest) parseSearchParam(queryParams map[string][]string) {
	if search := getQueryParam(queryParams, "search"); search != "" {
		r.Search = &search
	}
}

// parseFilterParams extracts category and boolean filter parameters
func (r *ProductFiltersRequest) parseFilterParams(queryParams map[string][]string) error {
	// Parse category_id
	if categoryIDStr := getQueryParam(queryParams, "category_id"); categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_category_id_format"}
		}
		r.CategoryID = &categoryID
	}

	// Parse boolean filters
	if err := r.parseBoolParam(queryParams, "is_active", &r.IsActive); err != nil {
		return err
	}
	if err := r.parseBoolParam(queryParams, "is_highlighted", &r.IsHighlighted); err != nil {
		return err
	}
	if err := r.parseBoolParam(queryParams, "is_promotional", &r.IsPromotional); err != nil {
		return err
	}

	return nil
}

// parseBoolParam parses a boolean query parameter into the target pointer
func (r *ProductFiltersRequest) parseBoolParam(queryParams map[string][]string, key string, target **bool) error {
	if str := getQueryParam(queryParams, key); str != "" {
		value, err := strconv.ParseBool(str)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_" + key + "_format"}
		}
		*target = &value
	}
	return nil
}

// parsePriceParams extracts price range parameters
func (r *ProductFiltersRequest) parsePriceParams(queryParams map[string][]string) error {
	if minPriceStr := getQueryParam(queryParams, "min_price"); minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_min_price_format"}
		}
		r.MinPrice = &minPrice
	}

	if maxPriceStr := getQueryParam(queryParams, "max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_max_price_format"}
		}
		r.MaxPrice = &maxPrice
	}

	return nil
}

// parsePaginationParams extracts sorting and pagination parameters
func (r *ProductFiltersRequest) parsePaginationParams(queryParams map[string][]string) error {
	r.SortBy = getQueryParam(queryParams, "sort")
	r.SortOrder = getQueryParam(queryParams, "order")

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		r.Limit = limit
	}

	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		r.Cursor = cursorStr
	}

	return nil
}

// getQueryParam safely retrieves a single query parameter value
func getQueryParam(queryParams map[string][]string, key string) string {
	values, exists := queryParams[key]
	if !exists || len(values) == 0 {
		return ""
	}
	return values[0]
}
