package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/pagination"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Shared Types
// =============================================================================

// CategoryRequest represents category data in HTTP requests.
type CategoryRequest struct {
	ID          int                   `json:"id,omitempty"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Image       *CategoryImageRequest `json:"image,omitempty"`
}

// CategoryImageRequest represents image data in category requests.
type CategoryImageRequest struct {
	ID  int    `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

// ToModel converts CategoryRequest to a domain model.
func (c *CategoryRequest) ToModel() *models.Category {
	category := &models.Category{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
	}
	if c.Image != nil {
		category.Image = &models.Image{
			ID:  c.Image.ID,
			URL: c.Image.URL,
		}
	}
	return category
}

// isValidCategoryImageType checks if the MIME type is valid for category images.
func isValidCategoryImageType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}
	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}

// =============================================================================
// Create
// =============================================================================

// CategoryCreateRequest represents the HTTP request for creating a category.
type CategoryCreateRequest struct {
	Category CategoryRequest       `json:"category"`
	ShopID   int                   `json:"shop_id"`
	Image    *multipart.FileHeader `json:"-"`
}

// NewCategoryCreateRequest creates a CategoryCreateRequest from an HTTP request.
// shopID comes from the JWT token context (injected by auth middleware).
func NewCategoryCreateRequest(r *http.Request, shopID int) (*CategoryCreateRequest, error) {
	// Extract category JSON from form data
	categoryJSON := r.FormValue("category")
	if strings.TrimSpace(categoryJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "category_json_required"}
	}

	// Parse category JSON
	var category CategoryRequest
	if err := json.Unmarshal([]byte(categoryJSON), &category); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_category_json_format"}
	}

	// Get image from form (single image, required)
	var imageHeader *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["image"]; exists && len(files) > 0 {
			imageHeader = files[0]
		}
	}
	if imageHeader == nil {
		return nil, &httpErrors.BadRequestError{Message: "category_image_required"}
	}

	return &CategoryCreateRequest{
		Category: category,
		ShopID:   shopID,
		Image:    imageHeader,
	}, nil
}

// ToModel converts the request's category to a domain model.
func (r *CategoryCreateRequest) ToModel() *models.Category {
	return r.Category.ToModel()
}

// Validate validates the HTTP request for creating a category.
// This validates HTTP-specific concerns (required fields, file format).
// Business validations are handled by Category.Validate() in the service layer.
func (r *CategoryCreateRequest) Validate() error {
	// Validate shop_id (required and positive)
	if r.ShopID <= 0 {
		return &httpErrors.BadRequestError{Message: "shop_id_is_required"}
	}

	// Validate category name (required field)
	if strings.TrimSpace(r.Category.Name) == "" {
		return &httpErrors.BadRequestError{Message: "category_name_is_required"}
	}

	// Validate category description (required field)
	if strings.TrimSpace(r.Category.Description) == "" {
		return &httpErrors.BadRequestError{Message: "category_description_is_required"}
	}

	// Validate image (size and type)
	if err := r.validateImage(); err != nil {
		return err
	}

	return nil
}

// validateImage validates the image file (size and MIME type).
// Note: Image presence is validated in the handler.
func (r *CategoryCreateRequest) validateImage() error {
	// Check file size (max 3MB)
	if r.Image.Size > 3*1024*1024 {
		return &httpErrors.BadRequestError{Message: "image_size_too_large_max_3mb"}
	}

	// Open file to check MIME type
	file, err := r.Image.Open()
	if err != nil {
		return &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	// Check MIME type
	mimeType := http.DetectContentType(buffer)
	if !isValidCategoryImageType(mimeType) {
		return &httpErrors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
	}

	// Reset file pointer for later use
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
		}
	}

	return nil
}

// ToImageBuffer converts the FileHeader to a byte slice for upload service.
func (r *CategoryCreateRequest) ToImageBuffer() ([]byte, error) {
	if r.Image == nil {
		return nil, nil
	}

	file, err := r.Image.Open()
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	return buffer.Bytes(), nil
}

// =============================================================================
// Filters
// =============================================================================

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
		cursorData, err := pagination.DecodeCursor(r.Cursor)
		if err != nil {
			// If cursor is invalid, treat as first page (graceful degradation)
			return filters
		}

		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}

// =============================================================================
// Update
// =============================================================================

// CategoryUpdateRequest represents the HTTP request for updating a category.
type CategoryUpdateRequest struct {
	Category CategoryRequest       `json:"category"`
	NewImage *multipart.FileHeader `json:"-"` // Optional new image to replace existing
}

// NewCategoryUpdateRequest creates a CategoryUpdateRequest from an HTTP request.
func NewCategoryUpdateRequest(r *http.Request) (*CategoryUpdateRequest, error) {
	// Extract category JSON from form data
	categoryJSON := r.FormValue("category")
	if strings.TrimSpace(categoryJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "category_json_required"}
	}

	// Parse category JSON
	var category CategoryRequest
	if err := json.Unmarshal([]byte(categoryJSON), &category); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_category_json_format"}
	}

	// Get new image from form (optional for update)
	var newImageHeader *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["image"]; exists && len(files) > 0 {
			newImageHeader = files[0]
		}
	}

	return &CategoryUpdateRequest{
		Category: category,
		NewImage: newImageHeader,
	}, nil
}

// ToModel converts the request's category to a domain model.
func (r *CategoryUpdateRequest) ToModel() *models.Category {
	return r.Category.ToModel()
}

// Validate validates the HTTP request for updating a category.
// This validates HTTP-specific concerns (required fields, file format).
// Business validations are handled by Category.Validate() in the service layer.
func (r *CategoryUpdateRequest) Validate() error {
	// Validate category name (required field)
	if strings.TrimSpace(r.Category.Name) == "" {
		return &httpErrors.BadRequestError{Message: "category_name_is_required"}
	}

	// Validate category description (required field)
	if strings.TrimSpace(r.Category.Description) == "" {
		return &httpErrors.BadRequestError{Message: "category_description_is_required"}
	}

	// CRITICAL: Validate that at least one image exists (existing OR new)
	// If no existing image and no new image, reject the request
	if r.Category.Image == nil && r.NewImage == nil {
		return &httpErrors.BadRequestError{Message: "category_image_required"}
	}

	// Validate existing image has valid data (if present and no new image)
	if r.Category.Image != nil && r.NewImage == nil {
		if err := r.validateExistingImage(); err != nil {
			return err
		}
	}

	// Validate new image (if provided)
	if r.NewImage != nil {
		if err := r.validateNewImage(); err != nil {
			return err
		}
	}

	return nil
}

// validateExistingImage validates that the existing image has valid ID and URL.
func (r *CategoryUpdateRequest) validateExistingImage() error {
	if r.Category.Image.ID <= 0 {
		return &httpErrors.BadRequestError{Message: "existing_image_must_have_valid_id"}
	}
	if strings.TrimSpace(r.Category.Image.URL) == "" {
		return &httpErrors.BadRequestError{Message: "existing_image_must_have_url"}
	}
	return nil
}

// validateNewImage validates the new image file (size and MIME type).
func (r *CategoryUpdateRequest) validateNewImage() error {
	// Check file size (max 3MB)
	if r.NewImage.Size > 3*1024*1024 {
		return &httpErrors.BadRequestError{Message: "image_size_too_large_max_3mb"}
	}

	// Open file to check MIME type
	file, err := r.NewImage.Open()
	if err != nil {
		return &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	// Check MIME type
	mimeType := http.DetectContentType(buffer)
	if !isValidCategoryImageType(mimeType) {
		return &httpErrors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
	}

	// Reset file pointer for later use
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
		}
	}

	return nil
}

// ToImageBuffer converts the new FileHeader to a byte slice for upload service.
// Returns nil if no new image is provided.
func (r *CategoryUpdateRequest) ToImageBuffer() ([]byte, error) {
	if r.NewImage == nil {
		return nil, nil
	}

	file, err := r.NewImage.Open()
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	return buffer.Bytes(), nil
}
