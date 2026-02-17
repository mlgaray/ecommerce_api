package requests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

// =============================================================================
// Multipart Helper
// =============================================================================

func createCategoryMultipartRequest(t *testing.T, categoryJSON string, imageBytes []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if categoryJSON != "" {
		_ = writer.WriteField("category", categoryJSON)
	}

	if imageBytes != nil {
		part, err := writer.CreateFormFile("image", "test.jpg")
		assert.NoError(t, err)
		_, err = part.Write(imageBytes)
		assert.NoError(t, err)
	}

	err := writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(7 << 20)
	assert.NoError(t, err)

	return req
}

// =============================================================================
// CategoryRequest.ToModel Tests
// =============================================================================

func TestCategoryRequest_ToModel(t *testing.T) {
	t.Run("when category with image then maps all fields", func(t *testing.T) {
		// Arrange
		request := &CategoryRequest{
			ID:          1,
			Name:        "Electronics",
			Description: "Electronic products",
			Image:       &CategoryImageRequest{ID: 10, URL: "https://example.com/img.jpg"},
		}

		// Act
		category := request.ToModel()

		// Assert
		assert.Equal(t, 1, category.ID)
		assert.Equal(t, "Electronics", category.Name)
		assert.Equal(t, "Electronic products", category.Description)
		assert.NotNil(t, category.Image)
		assert.Equal(t, 10, category.Image.ID)
		assert.Equal(t, "https://example.com/img.jpg", category.Image.URL)
	})

	t.Run("when category without image then image is nil", func(t *testing.T) {
		// Arrange
		request := &CategoryRequest{
			Name:        "Clothing",
			Description: "Clothing products",
		}

		// Act
		category := request.ToModel()

		// Assert
		assert.Equal(t, "Clothing", category.Name)
		assert.Nil(t, category.Image)
	})
}

// =============================================================================
// NewCategoryCreateRequest Tests
// =============================================================================

func TestNewCategoryCreateRequest(t *testing.T) {
	t.Run("when valid JSON and image then returns request", func(t *testing.T) {
		// Arrange
		categoryJSON := `{"name": "Electronics", "description": "Electronic products"}`
		req := createCategoryMultipartRequest(t, categoryJSON, createJPEGBytes())

		// Act
		result, err := NewCategoryCreateRequest(req, 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Electronics", result.Category.Name)
		assert.Equal(t, 1, result.ShopID)
		assert.NotNil(t, result.Image)
	})

	t.Run("when empty JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, "", createJPEGBytes())

		// Act
		result, err := NewCategoryCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_json_required", badReqErr.Message)
	})

	t.Run("when invalid JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, "not-json", createJPEGBytes())

		// Act
		result, err := NewCategoryCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_category_json_format", badReqErr.Message)
	})

	t.Run("when no image then returns error", func(t *testing.T) {
		// Arrange
		categoryJSON := `{"name": "Test", "description": "Test"}`
		req := createCategoryMultipartRequest(t, categoryJSON, nil)

		// Act
		result, err := NewCategoryCreateRequest(req, 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_image_required", badReqErr.Message)
	})
}

// =============================================================================
// CategoryCreateRequest.Validate Tests
// =============================================================================

func TestCategoryCreateRequest_Validate(t *testing.T) {
	t.Run("when valid request then no error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when shop_id is zero then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 0)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_id_is_required", badReqErr.Message)
	})

	t.Run("when name is empty then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_name_is_required", badReqErr.Message)
	})

	t.Run("when description is empty then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":""}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_description_is_required", badReqErr.Message)
	})

	t.Run("when image is valid JPEG then no error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when image is valid PNG then no error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createPNGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when image is invalid type then returns error", func(t *testing.T) {
		// Arrange
		invalidBytes := []byte("not an image file content that is long enough for detection")
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, invalidBytes)
		result, _ := NewCategoryCreateRequest(req, 1)

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
// CategoryCreateRequest.ToModel Tests
// =============================================================================

func TestCategoryCreateRequest_ToModel(t *testing.T) {
	t.Run("when create request then delegates to category ToModel", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Electronics","description":"Electronic products"}`, createJPEGBytes())
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		category := result.ToModel()

		// Assert
		assert.Equal(t, "Electronics", category.Name)
		assert.Equal(t, "Electronic products", category.Description)
	})
}

// =============================================================================
// CategoryCreateRequest.ToImageBuffer Tests
// =============================================================================

func TestCategoryCreateRequest_ToImageBuffer(t *testing.T) {
	t.Run("when image exists then returns bytes", func(t *testing.T) {
		// Arrange
		jpegBytes := createJPEGBytes()
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, jpegBytes)
		result, _ := NewCategoryCreateRequest(req, 1)

		// Act
		buffer, err := result.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, buffer)
		assert.Greater(t, len(buffer), 0)
	})

	t.Run("when no image then returns nil", func(t *testing.T) {
		// Arrange
		request := &CategoryCreateRequest{
			Category: CategoryRequest{Name: "Test"},
			Image:    nil,
		}

		// Act
		buffer, err := request.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, buffer)
	})
}

// =============================================================================
// NewCategoryUpdateRequest Tests
// =============================================================================

func TestNewCategoryUpdateRequest(t *testing.T) {
	t.Run("when valid JSON without new image then returns request", func(t *testing.T) {
		// Arrange
		categoryJSON := `{"name":"Updated","description":"Updated desc","image":{"id":1,"url":"https://example.com/img.jpg"}}`
		req := createCategoryMultipartRequest(t, categoryJSON, nil)

		// Act
		result, err := NewCategoryUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated", result.Category.Name)
		assert.NotNil(t, result.Category.Image)
		assert.Nil(t, result.NewImage)
	})

	t.Run("when valid JSON with new image then returns request with image", func(t *testing.T) {
		// Arrange
		categoryJSON := `{"name":"Updated","description":"Updated desc"}`
		req := createCategoryMultipartRequest(t, categoryJSON, createJPEGBytes())

		// Act
		result, err := NewCategoryUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.NewImage)
	})

	t.Run("when empty JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, "", nil)

		// Act
		result, err := NewCategoryUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_json_required", badReqErr.Message)
	})

	t.Run("when invalid JSON then returns error", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, "{invalid", nil)

		// Act
		result, err := NewCategoryUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_category_json_format", badReqErr.Message)
	})
}

// =============================================================================
// CategoryUpdateRequest.Validate Tests
// =============================================================================

func TestCategoryUpdateRequest_Validate(t *testing.T) {
	t.Run("when name is empty then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{Name: "", Description: "Desc", Image: &CategoryImageRequest{ID: 1, URL: "https://example.com/img.jpg"}},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_name_is_required", badReqErr.Message)
	})

	t.Run("when description is empty then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{Name: "Test", Description: "", Image: &CategoryImageRequest{ID: 1, URL: "https://example.com/img.jpg"}},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_description_is_required", badReqErr.Message)
	})

	t.Run("when no image at all then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{Name: "Test", Description: "Desc"},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "category_image_required", badReqErr.Message)
	})

	t.Run("when existing image with invalid ID then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Test",
				Description: "Desc",
				Image:       &CategoryImageRequest{ID: 0, URL: "https://example.com/img.jpg"},
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

	t.Run("when existing image with empty URL then returns error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Test",
				Description: "Desc",
				Image:       &CategoryImageRequest{ID: 1, URL: ""},
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

	t.Run("when existing image is valid then no error", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				Name:        "Test",
				Description: "Desc",
				Image:       &CategoryImageRequest{ID: 1, URL: "https://example.com/img.jpg"},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when new image provided then validates new image", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryUpdateRequest(req)

		// Act
		err := result.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when new image is invalid type then returns error", func(t *testing.T) {
		// Arrange
		invalidBytes := []byte("not an image content that is long enough for mime detection")
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, invalidBytes)
		result, _ := NewCategoryUpdateRequest(req)

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
// CategoryUpdateRequest.ToModel Tests
// =============================================================================

func TestCategoryUpdateRequest_ToModel(t *testing.T) {
	t.Run("when update request then delegates to category ToModel", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{
				ID:          5,
				Name:        "Updated",
				Description: "Updated desc",
			},
		}

		// Act
		category := request.ToModel()

		// Assert
		assert.Equal(t, 5, category.ID)
		assert.Equal(t, "Updated", category.Name)
	})
}

// =============================================================================
// CategoryUpdateRequest.ToImageBuffer Tests
// =============================================================================

func TestCategoryUpdateRequest_ToImageBuffer(t *testing.T) {
	t.Run("when no new image then returns nil", func(t *testing.T) {
		// Arrange
		request := &CategoryUpdateRequest{
			Category: CategoryRequest{Name: "Test"},
		}

		// Act
		buffer, err := request.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, buffer)
	})

	t.Run("when new image exists then returns bytes", func(t *testing.T) {
		// Arrange
		req := createCategoryMultipartRequest(t, `{"name":"Test","description":"Desc"}`, createJPEGBytes())
		result, _ := NewCategoryUpdateRequest(req)

		// Act
		buffer, err := result.ToImageBuffer()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, buffer)
		assert.Greater(t, len(buffer), 0)
	})
}

// =============================================================================
// isValidCategoryImageType Tests
// =============================================================================

func TestIsValidCategoryImageType(t *testing.T) {
	t.Run("when jpeg then valid", func(t *testing.T) {
		assert.True(t, isValidCategoryImageType("image/jpeg"))
	})

	t.Run("when jpg then valid", func(t *testing.T) {
		assert.True(t, isValidCategoryImageType("image/jpg"))
	})

	t.Run("when png then valid", func(t *testing.T) {
		assert.True(t, isValidCategoryImageType("image/png"))
	})

	t.Run("when gif then invalid", func(t *testing.T) {
		assert.False(t, isValidCategoryImageType("image/gif"))
	})

	t.Run("when text then invalid", func(t *testing.T) {
		assert.False(t, isValidCategoryImageType("text/plain"))
	})

	t.Run("when empty then invalid", func(t *testing.T) {
		assert.False(t, isValidCategoryImageType(""))
	})
}
