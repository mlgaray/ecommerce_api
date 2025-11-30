package contracts

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// CategoryCreateRequest represents the HTTP request for creating a category.
type CategoryCreateRequest struct {
	Category models.Category       `json:"category"`
	ShopID   int                   `json:"shop_id"`
	Image    *multipart.FileHeader `json:"-"`
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
