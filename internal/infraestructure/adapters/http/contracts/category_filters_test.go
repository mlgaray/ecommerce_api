package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// NewCategoryFiltersRequest Tests
// =============================================================================

func TestNewCategoryFiltersRequest(t *testing.T) {
	t.Run("when no params then returns empty request", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Search)
		assert.Equal(t, "", result.SortBy)
		assert.Equal(t, "", result.SortOrder)
		assert.Equal(t, 0, result.Limit)
		assert.Equal(t, "", result.Cursor)
	})

	t.Run("when search provided then sets search", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search": {"electronics"},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.Search)
		assert.Equal(t, "electronics", *result.Search)
	})

	t.Run("when sort and order provided then sets them", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"sort":  {"name"},
			"order": {"asc"},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "name", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
	})

	t.Run("when limit provided then sets limit", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"50"},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when invalid limit then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"abc"},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_limit_format", badReqErr.Message)
	})

	t.Run("when cursor provided then sets cursor", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"cursor": {"eyJsYXN0X2lkIjoxMH0="},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "eyJsYXN0X2lkIjoxMH0=", result.Cursor)
	})

	t.Run("when all params provided then sets all", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search": {"clothes"},
			"sort":   {"created_at"},
			"order":  {"desc"},
			"limit":  {"25"},
			"cursor": {"abc123"},
		}

		// Act
		result, err := NewCategoryFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "clothes", *result.Search)
		assert.Equal(t, "created_at", result.SortBy)
		assert.Equal(t, "desc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
		assert.Equal(t, "abc123", result.Cursor)
	})
}

// =============================================================================
// CategoryFiltersRequest.Validate Tests
// =============================================================================

func TestCategoryFiltersRequest_Validate(t *testing.T) {
	t.Run("when no filters then no error", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when search is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		search := "   "
		request := &CategoryFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "search_term_cannot_be_empty", badReqErr.Message)
	})

	t.Run("when search has content then trims and passes", func(t *testing.T) {
		// Arrange
		search := "  electronics  "
		request := &CategoryFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "electronics", *request.Search)
	})

	t.Run("when limit is negative then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{Limit: -5}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "limit_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when limit is zero then no error", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{Limit: 0}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when limit is positive then no error", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{Limit: 50}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when all valid then no error", func(t *testing.T) {
		// Arrange
		search := "electronics"
		request := &CategoryFiltersRequest{
			Search:    &search,
			SortBy:    "name",
			SortOrder: "asc",
			Limit:     20,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// CategoryFiltersRequest.ToCategoryFilters Tests
// =============================================================================

func TestCategoryFiltersRequest_ToCategoryFilters(t *testing.T) {
	t.Run("when no params then returns defaults", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{}

		// Act
		result := request.ToCategoryFilters()

		// Assert
		assert.Nil(t, result.Search)
		assert.Equal(t, "", result.SortBy)
		assert.Equal(t, "", result.SortOrder)
		assert.Equal(t, 0, result.Limit)
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when all params then maps all", func(t *testing.T) {
		// Arrange
		search := "electronics"
		request := &CategoryFiltersRequest{
			Search:    &search,
			SortBy:    "name",
			SortOrder: "asc",
			Limit:     25,
		}

		// Act
		result := request.ToCategoryFilters()

		// Assert
		assert.Equal(t, "electronics", *result.Search)
		assert.Equal(t, "name", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
	})

	t.Run("when invalid cursor then treats as first page", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{
			Cursor: "invalid_cursor_not_base64",
		}

		// Act
		result := request.ToCategoryFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when no cursor then LastID is nil", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{
			Cursor: "",
		}

		// Act
		result := request.ToCategoryFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when search is nil then result search is nil", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{
			Search: nil,
		}

		// Act
		result := request.ToCategoryFilters()

		// Assert
		assert.Nil(t, result.Search)
	})
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCategoryFiltersRequest_EdgeCases(t *testing.T) {
	t.Run("when search is single character then passes", func(t *testing.T) {
		// Arrange
		search := "a"
		request := &CategoryFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when search has leading and trailing whitespace then trims", func(t *testing.T) {
		// Arrange
		search := "\t  electronics  \n"
		request := &CategoryFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "electronics", *request.Search)
	})

	t.Run("when limit is very large then no error", func(t *testing.T) {
		// Arrange
		request := &CategoryFiltersRequest{Limit: 999999}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when sort contains special characters then passes", func(t *testing.T) {
		// Arrange - HTTP layer doesn't validate sort field values
		request := &CategoryFiltersRequest{SortBy: "invalid_field"}

		// Act
		err := request.Validate()

		// Assert
		// HTTP layer accepts any string - business validation happens in domain layer
		assert.NoError(t, err)
	})
}
