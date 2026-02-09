package contracts

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

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
