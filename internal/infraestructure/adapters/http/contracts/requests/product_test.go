package requests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// NewProductFiltersRequest Tests
// =============================================================================

func TestNewProductFiltersRequest(t *testing.T) {
	t.Run("when no params then returns empty request", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Search)
		assert.Nil(t, result.CategoryID)
		assert.Equal(t, 0, result.Limit)
	})

	t.Run("when search provided then sets search", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search": {"laptop"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.Search)
		assert.Equal(t, "laptop", *result.Search)
	})

	t.Run("when category_id provided then sets category", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"category_id": {"5"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.CategoryID)
		assert.Equal(t, 5, *result.CategoryID)
	})

	t.Run("when invalid category_id then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"category_id": {"abc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_category_id_format", badReqErr.Message)
	})

	t.Run("when is_active true provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.True(t, *result.IsActive)
	})

	t.Run("when is_active false provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"false"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("when invalid is_active then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"maybe"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_is_active_format", badReqErr.Message)
	})

	t.Run("when is_highlighted provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_highlighted": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsHighlighted)
		assert.True(t, *result.IsHighlighted)
	})

	t.Run("when is_promotional provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_promotional": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsPromotional)
		assert.True(t, *result.IsPromotional)
	})

	t.Run("when min_price provided then sets price", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"min_price": {"100.50"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.MinPrice)
		assert.Equal(t, 100.50, *result.MinPrice)
	})

	t.Run("when invalid min_price then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"min_price": {"abc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_min_price_format", badReqErr.Message)
	})

	t.Run("when max_price provided then sets price", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"max_price": {"500.99"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.MaxPrice)
		assert.Equal(t, 500.99, *result.MaxPrice)
	})

	t.Run("when invalid max_price then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"max_price": {"xyz"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_max_price_format", badReqErr.Message)
	})

	t.Run("when limit provided then sets limit", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"50"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

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
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_limit_format", badReqErr.Message)
	})

	t.Run("when sort and order provided then sets them", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"sort":  {"price"},
			"order": {"asc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "price", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
	})

	t.Run("when cursor provided then sets cursor", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"cursor": {"eyJsYXN0X2lkIjoxMH0="},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "eyJsYXN0X2lkIjoxMH0=", result.Cursor)
	})

	t.Run("when all params provided then sets all", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search":         {"laptop"},
			"category_id":    {"5"},
			"is_active":      {"true"},
			"is_highlighted": {"false"},
			"is_promotional": {"true"},
			"min_price":      {"100"},
			"max_price":      {"500"},
			"sort":           {"name"},
			"order":          {"desc"},
			"limit":          {"25"},
			"cursor":         {"abc123"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "laptop", *result.Search)
		assert.Equal(t, 5, *result.CategoryID)
		assert.True(t, *result.IsActive)
		assert.False(t, *result.IsHighlighted)
		assert.True(t, *result.IsPromotional)
		assert.Equal(t, 100.0, *result.MinPrice)
		assert.Equal(t, 500.0, *result.MaxPrice)
		assert.Equal(t, "name", result.SortBy)
		assert.Equal(t, "desc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
		assert.Equal(t, "abc123", result.Cursor)
	})
}

// =============================================================================
// ProductFiltersRequest.Validate Tests
// =============================================================================

func TestProductFiltersRequest_Validate(t *testing.T) {
	t.Run("when no filters then no error", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when search is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		search := "   "
		request := &ProductFiltersRequest{Search: &search}

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
		search := "  laptop  "
		request := &ProductFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "laptop", *request.Search)
	})

	t.Run("when limit is negative then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{Limit: -5}

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
		request := &ProductFiltersRequest{Limit: 0}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when min_price is negative then returns error", func(t *testing.T) {
		// Arrange
		minPrice := -10.0
		request := &ProductFiltersRequest{MinPrice: &minPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "min_price_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when max_price is negative then returns error", func(t *testing.T) {
		// Arrange
		maxPrice := -10.0
		request := &ProductFiltersRequest{MaxPrice: &maxPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "max_price_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when min_price is zero then no error", func(t *testing.T) {
		// Arrange
		minPrice := 0.0
		request := &ProductFiltersRequest{MinPrice: &minPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when all valid then no error", func(t *testing.T) {
		// Arrange
		search := "laptop"
		minPrice := 100.0
		maxPrice := 500.0
		request := &ProductFiltersRequest{
			Search:   &search,
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
			Limit:    20,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// ProductFiltersRequest.ToProductFilters Tests
// =============================================================================

func TestProductFiltersRequest_ToProductFilters(t *testing.T) {
	t.Run("when no params then returns defaults", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.Search)
		assert.Nil(t, result.CategoryID)
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when all params then maps all", func(t *testing.T) {
		// Arrange
		search := "laptop"
		categoryID := 5
		isActive := true
		minPrice := 100.0
		maxPrice := 500.0
		request := &ProductFiltersRequest{
			Search:     &search,
			CategoryID: &categoryID,
			IsActive:   &isActive,
			MinPrice:   &minPrice,
			MaxPrice:   &maxPrice,
			SortBy:     "price",
			SortOrder:  "asc",
			Limit:      25,
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Equal(t, "laptop", *result.Search)
		assert.Equal(t, 5, *result.CategoryID)
		assert.True(t, *result.IsActive)
		assert.Equal(t, 100.0, *result.MinPrice)
		assert.Equal(t, 500.0, *result.MaxPrice)
		assert.Equal(t, "price", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
	})

	t.Run("when invalid cursor then treats as first page", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{
			Cursor: "invalid_cursor_not_base64",
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when no cursor then LastID is nil", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{
			Cursor: "",
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})
}

// =============================================================================
// getQueryParam Tests
// =============================================================================

func TestGetQueryParam(t *testing.T) {
	t.Run("when key exists then returns value", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {"value"},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "value", result)
	})

	t.Run("when key not exists then returns empty", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("when key exists with empty array then returns empty", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("when key has multiple values then returns first", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {"first", "second", "third"},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "first", result)
	})
}

// =============================================================================
// Product Multipart Helper
// =============================================================================

func createProductMultipartRequest(t *testing.T, productJSON string, imageFiles [][]byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if productJSON != "" {
		_ = writer.WriteField("product", productJSON)
	}

	for i, imgBytes := range imageFiles {
		key := "images[" + strconv.Itoa(i) + "]"
		part, err := writer.CreateFormFile(key, "image"+strconv.Itoa(i)+".jpg")
		assert.NoError(t, err)
		_, err = part.Write(imgBytes)
		assert.NoError(t, err)
	}

	err := writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/products", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(7 << 20)
	assert.NoError(t, err)

	return req
}

func validProductJSON() string {
	return `{
		"name": "Laptop",
		"description": "A powerful laptop",
		"price": 999.99,
		"category": {"id": 1},
		"is_active": true,
		"variants": [
			{
				"name": "Color",
				"selection_type": "single",
				"options": [{"name": "Silver", "price": 0}]
			}
		]
	}`
}

// =============================================================================
// ProductRequest.ToModel Tests
// =============================================================================

func TestProductRequest_ToModel(t *testing.T) {
	t.Run("when full product then maps all fields", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{
			ID:               1,
			Name:             "Laptop",
			Description:      "A powerful laptop",
			Price:            999.99,
			IsActive:         true,
			IsPromotional:    true,
			PromotionalPrice: 799.99,
			IsHighlighted:    true,
			IsStockeable:     true,
			Stock:            50,
			MinimumStock:     5,
			Category:         &ProductCategoryRequest{ID: 1},
			Variants: []*ProductVariantRequest{
				{
					ID:            10,
					Name:          "Color",
					Order:         1,
					SelectionType: "single",
					MaxSelections: 1,
					IsRequired:    true,
					Options: []*ProductOptionRequest{
						{ID: 100, Name: "Silver", Price: 0, Order: 1},
						{ID: 101, Name: "Gold", Price: 50, Order: 2},
					},
				},
			},
			Images: []*ProductImageRequest{
				{ID: 1, URL: "https://example.com/img1.jpg"},
				{ID: 2, URL: "https://example.com/img2.jpg"},
			},
		}

		// Act
		product := request.ToModel()

		// Assert
		assert.Equal(t, 1, product.ID)
		assert.Equal(t, "Laptop", product.Name)
		assert.Equal(t, "A powerful laptop", product.Description)
		assert.Equal(t, 999.99, product.Price)
		assert.True(t, product.IsActive)
		assert.True(t, product.IsPromotional)
		assert.Equal(t, 799.99, product.PromotionalPrice)
		assert.True(t, product.IsHighlighted)
		assert.True(t, product.IsStockeable)
		assert.Equal(t, 50, product.Stock)
		assert.Equal(t, 5, product.MinimumStock)
	})

	t.Run("when product has category then maps category ID", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{
			Name:     "Test",
			Category: &ProductCategoryRequest{ID: 5},
		}

		// Act
		product := request.ToModel()

		// Assert
		assert.NotNil(t, product.Category)
		assert.Equal(t, 5, product.Category.ID)
	})

	t.Run("when product has no category then category is nil", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{Name: "Test"}

		// Act
		product := request.ToModel()

		// Assert
		assert.Nil(t, product.Category)
	})

	t.Run("when product has variants then maps variants with options", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{
			Name: "Test",
			Variants: []*ProductVariantRequest{
				{
					ID:            10,
					Name:          "Size",
					Order:         1,
					SelectionType: "single",
					MaxSelections: 1,
					IsRequired:    true,
					Options: []*ProductOptionRequest{
						{ID: 100, Name: "Small", Price: 0, Order: 1},
						{ID: 101, Name: "Large", Price: 5, Order: 2},
					},
				},
			},
		}

		// Act
		product := request.ToModel()

		// Assert
		assert.Len(t, product.Variants, 1)
		assert.Equal(t, 10, product.Variants[0].ID)
		assert.Equal(t, "Size", product.Variants[0].Name)
		assert.Equal(t, models.SelectionType("single"), product.Variants[0].SelectionType)
		assert.Len(t, product.Variants[0].Options, 2)
		assert.Equal(t, "Small", product.Variants[0].Options[0].Name)
	})

	t.Run("when product has images then maps images", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{
			Name: "Test",
			Images: []*ProductImageRequest{
				{ID: 1, URL: "https://example.com/img.jpg"},
			},
		}

		// Act
		product := request.ToModel()

		// Assert
		assert.Len(t, product.Images, 1)
		assert.Equal(t, 1, product.Images[0].ID)
		assert.Equal(t, "https://example.com/img.jpg", product.Images[0].URL)
	})

	t.Run("when product has no optional fields then nil", func(t *testing.T) {
		// Arrange
		request := &ProductRequest{Name: "Minimal"}

		// Act
		product := request.ToModel()

		// Assert
		assert.Nil(t, product.Category)
		assert.Nil(t, product.Variants)
		assert.Nil(t, product.Images)
	})
}

// =============================================================================
// ProductVariantRequest.ToModel Tests
// =============================================================================

func TestProductVariantRequest_ToModel(t *testing.T) {
	t.Run("when variant with options then maps all", func(t *testing.T) {
		// Arrange
		request := &ProductVariantRequest{
			ID:            1,
			Name:          "Toppings",
			Order:         1,
			SelectionType: "multiple",
			MaxSelections: 3,
			IsRequired:    false,
			Options: []*ProductOptionRequest{
				{ID: 10, Name: "Cheese", Price: 1.50, Order: 1},
				{ID: 11, Name: "Bacon", Price: 2.00, Order: 2},
			},
		}

		// Act
		variant := request.ToModel()

		// Assert
		assert.Equal(t, 1, variant.ID)
		assert.Equal(t, "Toppings", variant.Name)
		assert.Equal(t, 1, variant.Order)
		assert.Equal(t, models.SelectionType("multiple"), variant.SelectionType)
		assert.Equal(t, 3, variant.MaxSelections)
		assert.False(t, variant.IsRequired)
		assert.Len(t, variant.Options, 2)
		assert.Equal(t, "Cheese", variant.Options[0].Name)
		assert.Equal(t, 1.50, variant.Options[0].Price)
	})

	t.Run("when variant with no options then options is nil", func(t *testing.T) {
		// Arrange
		request := &ProductVariantRequest{
			Name:          "Size",
			SelectionType: "single",
		}

		// Act
		variant := request.ToModel()

		// Assert
		assert.Nil(t, variant.Options)
	})
}

// =============================================================================
// NewProductCreateRequest Tests
// =============================================================================

func TestNewProductCreateRequest(t *testing.T) {
	t.Run("when valid JSON and images then returns request", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{createJPEGBytes()})

		// Act
		result, err := NewProductCreateRequest(req, 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Laptop", result.Product.Name)
		assert.Equal(t, 1, result.ShopID)
		assert.Len(t, result.Images, 1)
	})

	t.Run("when empty JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, "", [][]byte{createJPEGBytes()})

		// Act
		result, err := NewProductCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_json_required", badReqErr.Message)
	})

	t.Run("when invalid JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, "{invalid", [][]byte{createJPEGBytes()})

		// Act
		result, err := NewProductCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_product_json_format", badReqErr.Message)
	})

	t.Run("when no images then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), nil)

		// Act
		result, err := NewProductCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_image_required", badReqErr.Message)
	})

	t.Run("when multiple images then includes all", func(t *testing.T) {
		// Arrange
		images := [][]byte{createJPEGBytes(), createPNGBytes()}
		req := createProductMultipartRequest(t, validProductJSON(), images)

		// Act
		result, err := NewProductCreateRequest(req, 1)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result.Images, 2)
	})
}

// =============================================================================
// ProductCreateRequest.Validate Tests
// =============================================================================

func TestProductCreateRequest_Validate(t *testing.T) {
	t.Run("when valid request then no error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when name is empty then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"","description":"Desc","price":10,"category":{"id":1}}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_name_is_required", badReqErr.Message)
	})

	t.Run("when description is empty then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"","price":10,"category":{"id":1}}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_description_is_required", badReqErr.Message)
	})

	t.Run("when category is nil then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_id_is_required", badReqErr.Message)
	})

	t.Run("when category id is zero then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":0}}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_id_is_required", badReqErr.Message)
	})

	t.Run("when shop_id is zero then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 0)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_id_is_required", badReqErr.Message)
	})

	t.Run("when variant name is empty then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"","selection_type":"single","options":[{"name":"A"}]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "variant_name_is_required", badReqErr.Message)
	})

	t.Run("when variant selection_type is empty then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"Color","selection_type":"","options":[{"name":"A"}]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "variant_selection_type_is_required", badReqErr.Message)
	})

	t.Run("when variant selection_type is invalid then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"Color","selection_type":"invalid","options":[{"name":"A"}]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_selection_type_must_be_single_multiple_or_custom", badReqErr.Message)
	})

	t.Run("when variant has no options then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"Color","selection_type":"single","options":[]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "variant_must_have_at_least_one_option", badReqErr.Message)
	})

	t.Run("when option name is empty then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"Color","selection_type":"single","options":[{"name":""}]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "option_name_is_required", badReqErr.Message)
	})

	t.Run("when option price is negative then returns error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1},"variants":[{"name":"Color","selection_type":"single","options":[{"name":"Red","price":-1}]}]}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "option_price_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when image is invalid type then returns error", func(t *testing.T) {
		// Arrange
		invalidBytes := []byte("not an image content long enough for detection purpose test")
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{invalidBytes})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_image_type_only_jpeg_png_allowed", badReqErr.Message)
	})
}

// =============================================================================
// ProductCreateRequest.ToModel Tests
// =============================================================================

func TestProductCreateRequest_ToModel(t *testing.T) {
	t.Run("when create request then delegates to product ToModel", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{createJPEGBytes()})
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		product := result.ToModel()

		// Assert
		assert.Equal(t, "Laptop", product.Name)
		assert.Equal(t, 999.99, product.Price)
	})
}

// =============================================================================
// ProductCreateRequest.ToImageBuffers Tests
// =============================================================================

func TestProductCreateRequest_ToImageBuffers(t *testing.T) {
	t.Run("when images exist then returns byte slices", func(t *testing.T) {
		// Arrange
		images := [][]byte{createJPEGBytes(), createPNGBytes()}
		req := createProductMultipartRequest(t, validProductJSON(), images)
		result, _ := NewProductCreateRequest(req, 1)

		// Act
		buffers, err := result.ToImageBuffers()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, buffers, 2)
		assert.Greater(t, len(buffers[0]), 0)
		assert.Greater(t, len(buffers[1]), 0)
	})
}

// =============================================================================
// NewProductUpdateRequest Tests
// =============================================================================

func TestNewProductUpdateRequest(t *testing.T) {
	t.Run("when valid JSON without images then returns request", func(t *testing.T) {
		// Arrange
		json := `{"name":"Updated","description":"Updated desc","price":10,"category":{"id":1},"images":[{"id":1,"url":"https://example.com/img.jpg"}]}`
		req := createProductMultipartRequest(t, json, nil)

		// Act
		result, err := NewProductUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated", result.Product.Name)
		assert.Len(t, result.NewImages, 0)
	})

	t.Run("when valid JSON with new images then includes images", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, validProductJSON(), [][]byte{createJPEGBytes()})

		// Act
		result, err := NewProductUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.NewImages, 1)
	})

	t.Run("when empty JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, "", nil)

		// Act
		result, err := NewProductUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_json_required", badReqErr.Message)
	})

	t.Run("when invalid JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createProductMultipartRequest(t, "{bad json", nil)

		// Act
		result, err := NewProductUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_product_json_format", badReqErr.Message)
	})
}

// =============================================================================
// ProductUpdateRequest.Validate Tests
// =============================================================================

func TestProductUpdateRequest_Validate(t *testing.T) {
	t.Run("when valid with existing images then no error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "Test",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
				Images:      []*ProductImageRequest{{ID: 1, URL: "https://example.com/img.jpg"}},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when name is empty then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
				Images:      []*ProductImageRequest{{ID: 1, URL: "https://example.com/img.jpg"}},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_name_is_required", badReqErr.Message)
	})

	t.Run("when no images at all then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "Test",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "product_image_required", badReqErr.Message)
	})

	t.Run("when existing image has invalid ID then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "Test",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
				Images:      []*ProductImageRequest{{ID: 0, URL: "https://example.com/img.jpg"}},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_valid_id", badReqErr.Message)
	})

	t.Run("when existing image has empty URL then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "Test",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
				Images:      []*ProductImageRequest{{ID: 1, URL: ""}},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_url", badReqErr.Message)
	})

	t.Run("when new images valid JPEG then no error", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1}}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes()})
		result, _ := NewProductUpdateRequest(req)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when new image is invalid type then returns error", func(t *testing.T) {
		// Arrange
		invalidBytes := []byte("not an image content that is long enough for detection purposes test")
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1}}`
		req := createProductMultipartRequest(t, json, [][]byte{invalidBytes})
		result, _ := NewProductUpdateRequest(req)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_image_type_only_jpeg_png_allowed", badReqErr.Message)
	})

	t.Run("when variant validation fails then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				Name:        "Test",
				Description: "Desc",
				Category:    &ProductCategoryRequest{ID: 1},
				Images:      []*ProductImageRequest{{ID: 1, URL: "https://example.com/img.jpg"}},
				Variants: []*ProductVariantRequest{
					{Name: "", SelectionType: "single", Options: []*ProductOptionRequest{{Name: "A"}}},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "variant_name_is_required", badReqErr.Message)
	})
}

// =============================================================================
// ProductUpdateRequest.ToModel Tests
// =============================================================================

func TestProductUpdateRequest_ToModel(t *testing.T) {
	t.Run("when update request then delegates to product ToModel", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{
				ID:   5,
				Name: "Updated Product",
			},
		}

		// Act
		product := request.ToModel()

		// Assert
		assert.Equal(t, 5, product.ID)
		assert.Equal(t, "Updated Product", product.Name)
	})
}

// =============================================================================
// ProductUpdateRequest.ToImageBuffers Tests
// =============================================================================

func TestProductUpdateRequest_ToImageBuffers(t *testing.T) {
	t.Run("when new images exist then returns byte slices", func(t *testing.T) {
		// Arrange
		json := `{"name":"Test","description":"Desc","price":10,"category":{"id":1}}`
		req := createProductMultipartRequest(t, json, [][]byte{createJPEGBytes(), createPNGBytes()})
		result, _ := NewProductUpdateRequest(req)

		// Act
		buffers, err := result.ToImageBuffers()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, buffers, 2)
	})

	t.Run("when no new images then returns empty slice", func(t *testing.T) {
		// Arrange
		request := &ProductUpdateRequest{
			Product: ProductRequest{Name: "Test"},
		}

		// Act
		buffers, err := request.ToImageBuffers()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, buffers, 0)
	})
}

// =============================================================================
// isValidImageType Tests
// =============================================================================

func TestIsValidImageType(t *testing.T) {
	t.Run("when jpeg then valid", func(t *testing.T) {
		assert.True(t, isValidImageType("image/jpeg"))
	})

	t.Run("when jpg then valid", func(t *testing.T) {
		assert.True(t, isValidImageType("image/jpg"))
	})

	t.Run("when png then valid", func(t *testing.T) {
		assert.True(t, isValidImageType("image/png"))
	})

	t.Run("when gif then invalid", func(t *testing.T) {
		assert.False(t, isValidImageType("image/gif"))
	})

	t.Run("when empty then invalid", func(t *testing.T) {
		assert.False(t, isValidImageType(""))
	})
}

// =============================================================================
// isValidSelectionType Tests
// =============================================================================

func TestIsValidSelectionType(t *testing.T) {
	t.Run("when single then valid", func(t *testing.T) {
		assert.True(t, isValidSelectionType("single"))
	})

	t.Run("when multiple then valid", func(t *testing.T) {
		assert.True(t, isValidSelectionType("multiple"))
	})

	t.Run("when custom then valid", func(t *testing.T) {
		assert.True(t, isValidSelectionType("custom"))
	})

	t.Run("when invalid then false", func(t *testing.T) {
		assert.False(t, isValidSelectionType("dropdown"))
	})

	t.Run("when empty then false", func(t *testing.T) {
		assert.False(t, isValidSelectionType(""))
	})
}
