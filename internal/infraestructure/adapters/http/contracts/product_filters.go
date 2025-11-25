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
type ProductFiltersRequest struct {
	ShopID        int      // From URL path parameter
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
func (r *ProductFiltersRequest) Validate() error {
	// HTTP validation: shop_id is required (comes from URL path)
	if r.ShopID <= 0 {
		return &httpErrors.BadRequestError{Message: "shop_id_is_required_in_path"}
	}

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
	// Decoding happens in ToModel() method - invalid cursors are treated as first page

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

// ToModel converts HTTP request to domain model
// Decodes the opaque cursor (if present) into LastID and LastSortValue
func (r *ProductFiltersRequest) ToModel() models.ProductFilters {
	filters := models.ProductFilters{
		ShopID:        r.ShopID,
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
		LastID:        nil,        // Will be populated from cursor
		LastSortValue: nil,        // Will be populated from cursor
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

// ParseQueryParams parses query parameters from URL into ProductFiltersRequest
// This helper function extracts and converts HTTP query params
func ParseQueryParams(queryParams map[string][]string, shopID int) (*ProductFiltersRequest, error) {
	request := &ProductFiltersRequest{
		ShopID: shopID,
	}

	// Parse search
	if search := getQueryParam(queryParams, "search"); search != "" {
		request.Search = &search
	}

	// Parse category_id
	if categoryIDStr := getQueryParam(queryParams, "category_id"); categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_category_id_format"}
		}
		request.CategoryID = &categoryID
	}

	// Parse is_active
	if isActiveStr := getQueryParam(queryParams, "is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_is_active_format"}
		}
		request.IsActive = &isActive
	}

	// Parse is_highlighted
	if isHighlightedStr := getQueryParam(queryParams, "is_highlighted"); isHighlightedStr != "" {
		isHighlighted, err := strconv.ParseBool(isHighlightedStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_is_highlighted_format"}
		}
		request.IsHighlighted = &isHighlighted
	}

	// Parse is_promotional
	if isPromotionalStr := getQueryParam(queryParams, "is_promotional"); isPromotionalStr != "" {
		isPromotional, err := strconv.ParseBool(isPromotionalStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_is_promotional_format"}
		}
		request.IsPromotional = &isPromotional
	}

	// Parse min_price
	if minPriceStr := getQueryParam(queryParams, "min_price"); minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_min_price_format"}
		}
		request.MinPrice = &minPrice
	}

	// Parse max_price
	if maxPriceStr := getQueryParam(queryParams, "max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_max_price_format"}
		}
		request.MaxPrice = &maxPrice
	}

	// Parse sort
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

	// Parse cursor (opaque string, no parsing needed)
	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		request.Cursor = cursorStr
	}

	return request, nil
}

// getQueryParam safely retrieves a single query parameter value
func getQueryParam(queryParams map[string][]string, key string) string {
	values, exists := queryParams[key]
	if !exists || len(values) == 0 {
		return ""
	}
	return values[0]
}
