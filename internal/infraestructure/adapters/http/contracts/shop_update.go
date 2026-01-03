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

// ShopUpdateRequest represents the multipart form request for updating a shop.
// Form fields:
// - "shop": JSON string with shop data (required)
// - "logo": optional new logo image file
// - "cover": optional new cover image file
type ShopUpdateRequest struct {
	Shop     models.Shop           `json:"shop"`
	NewLogo  *multipart.FileHeader `json:"-"` // Optional new logo image
	NewCover *multipart.FileHeader `json:"-"` // Optional new cover image
}

// NewShopUpdateRequest creates a ShopUpdateRequest from an HTTP request.
//
//nolint:gocyclo // Complexity is intentional for readability - multipart form parsing
func NewShopUpdateRequest(r *http.Request) (*ShopUpdateRequest, error) {
	// Extract shop JSON from form data
	shopJSON := r.FormValue("shop")
	if strings.TrimSpace(shopJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "shop_json_required"}
	}

	// Parse shop JSON (includes existing images with IDs)
	var shop models.Shop
	if err := json.Unmarshal([]byte(shopJSON), &shop); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_shop_json_format"}
	}

	// Get new logo from multipart form (optional)
	var newLogo *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["logo"]; exists && len(files) > 0 {
			newLogo = files[0]
		}
	}

	// Get new cover from multipart form (optional)
	var newCover *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["cover"]; exists && len(files) > 0 {
			newCover = files[0]
		}
	}

	return &ShopUpdateRequest{
		Shop:     shop,
		NewLogo:  newLogo,
		NewCover: newCover,
	}, nil
}

// Validate validates the shop update request.
func (r *ShopUpdateRequest) Validate() error {
	// Validate basic shop fields
	if err := r.validateBasicFields(); err != nil {
		return err
	}

	// Validate existing images have valid data
	if err := r.validateExistingImages(); err != nil {
		return err
	}

	// Validate new images (if any)
	if r.NewLogo != nil {
		if err := r.validateImage(r.NewLogo, "logo"); err != nil {
			return err
		}
	}
	if r.NewCover != nil {
		if err := r.validateImage(r.NewCover, "cover"); err != nil {
			return err
		}
	}

	return nil
}

func (r *ShopUpdateRequest) validateBasicFields() error {
	// Shop name is required
	if strings.TrimSpace(r.Shop.Name) == "" {
		return &httpErrors.BadRequestError{Message: "shop_name_is_required"}
	}

	// Shop slug is required
	if strings.TrimSpace(r.Shop.Slug) == "" {
		return &httpErrors.BadRequestError{Message: "shop_slug_is_required"}
	}

	return nil
}

func (r *ShopUpdateRequest) validateExistingImages() error {
	// Validate that existing images have valid IDs and URLs
	for _, img := range r.Shop.Images {
		if img.ID <= 0 {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_valid_id"}
		}
		if strings.TrimSpace(img.URL) == "" {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_url"}
		}
		// Validate type is logo or cover
		if img.Type != "logo" && img.Type != "cover" {
			return &httpErrors.BadRequestError{Message: "image_type_must_be_logo_or_cover"}
		}
	}
	return nil
}

func (r *ShopUpdateRequest) validateImage(imageHeader *multipart.FileHeader, imageType string) error {
	// Check file size (max 3MB per image)
	if imageHeader.Size > 3*1024*1024 {
		return &httpErrors.BadRequestError{Message: imageType + "_image_size_too_large_max_3mb"}
	}

	// Open file to check MIME type
	file, err := imageHeader.Open()
	if err != nil {
		return &httpErrors.BadRequestError{Message: "cannot_open_" + imageType + "_image_file"}
	}
	defer file.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return &httpErrors.BadRequestError{Message: "cannot_read_" + imageType + "_image_file"}
	}

	// Check MIME type
	mimeType := http.DetectContentType(buffer)
	if !isValidImageType(mimeType) {
		return &httpErrors.BadRequestError{Message: "invalid_" + imageType + "_image_type_only_jpeg_png_allowed"}
	}

	// Reset file pointer for later use
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
		}
	}

	return nil
}

// ToLogoBuffer converts the logo FileHeader to byte slice for upload service.
// Returns nil if no new logo was provided.
func (r *ShopUpdateRequest) ToLogoBuffer() ([]byte, error) {
	if r.NewLogo == nil {
		return nil, nil
	}
	return r.fileHeaderToBuffer(r.NewLogo)
}

// ToCoverBuffer converts the cover FileHeader to byte slice for upload service.
// Returns nil if no new cover was provided.
func (r *ShopUpdateRequest) ToCoverBuffer() ([]byte, error) {
	if r.NewCover == nil {
		return nil, nil
	}
	return r.fileHeaderToBuffer(r.NewCover)
}

func (r *ShopUpdateRequest) fileHeaderToBuffer(imageHeader *multipart.FileHeader) ([]byte, error) {
	file, err := imageHeader.Open()
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
