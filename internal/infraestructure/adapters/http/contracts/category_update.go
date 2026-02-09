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
