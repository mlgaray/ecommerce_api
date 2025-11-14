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

type ProductUpdateRequest struct {
	Product   models.Product          `json:"product"`
	ShopID    int                     `json:"shop_id"`
	NewImages []*multipart.FileHeader `json:"-"` // Optional new images to upload
}

func (r *ProductUpdateRequest) Validate() error {
	// Validate product data
	if err := r.validateProduct(); err != nil {
		return err
	}

	// Validate shop ID (optional for security checks)
	if r.ShopID <= 0 {
		return &httpErrors.BadRequestError{Message: "shop_id_is_required"}
	}

	// CRITICAL: Validate that at least one image exists (existing OR new)
	// User could delete all existing images, so we need at least one new image
	if len(r.Product.Images) == 0 && len(r.NewImages) == 0 {
		return &httpErrors.BadRequestError{Message: "at_least_one_image_is_required"}
	}

	// Validate new images (if any)
	if len(r.NewImages) > 0 {
		if err := r.validateNewImages(); err != nil {
			return err
		}
	}

	// Validate existing images have valid data
	if err := r.validateExistingImages(); err != nil {
		return err
	}

	return nil
}

func (r *ProductUpdateRequest) validateProduct() error {
	if err := r.validateBasicProductFields(); err != nil {
		return err
	}

	if err := r.validateVariants(); err != nil {
		return err
	}

	// Note: Product.Validate() will be called in the service layer
	// to validate business rules (price, stock, promotional price, etc.)
	return nil
}

func (r *ProductUpdateRequest) validateBasicProductFields() error {
	// HTTP validation: required fields
	if strings.TrimSpace(r.Product.Name) == "" {
		return &httpErrors.BadRequestError{Message: "product_name_is_required"}
	}
	if strings.TrimSpace(r.Product.Description) == "" {
		return &httpErrors.BadRequestError{Message: "product_description_is_required"}
	}
	if r.Product.Category == nil || r.Product.Category.ID <= 0 {
		return &httpErrors.BadRequestError{Message: "category_id_is_required"}
	}

	// Note: Business validations (price, stock, promotional price, etc.)
	// are handled by Product.Validate() in the service layer
	return nil
}

func (r *ProductUpdateRequest) validateVariants() error {
	for i, variant := range r.Product.Variants {
		if strings.TrimSpace(variant.Name) == "" {
			return &httpErrors.BadRequestError{Message: "variant_name_is_required"}
		}
		if variant.SelectionType == "" {
			return &httpErrors.BadRequestError{Message: "variant_selection_type_is_required"}
		}
		// Validate selection type is one of the allowed values
		if !isValidSelectionType(variant.SelectionType) {
			return &httpErrors.BadRequestError{Message: "invalid_selection_type_must_be_single_multiple_or_custom"}
		}
		if len(variant.Options) == 0 {
			return &httpErrors.BadRequestError{Message: "variant_must_have_at_least_one_option"}
		}

		if err := r.validateVariantOptions(variant, i); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductUpdateRequest) validateVariantOptions(variant *models.Variant, variantIndex int) error {
	for j, option := range variant.Options {
		if strings.TrimSpace(option.Name) == "" {
			return &httpErrors.BadRequestError{Message: "option_name_is_required"}
		}
		if option.Price < 0 {
			return &httpErrors.BadRequestError{Message: "option_price_cannot_be_negative"}
		}
		if option.Order == 0 {
			option.Order = j
		}
	}
	if variant.Order == 0 {
		variant.Order = variantIndex
	}
	return nil
}

func (r *ProductUpdateRequest) validateExistingImages() error {
	// Validate that existing images have valid IDs and URLs
	for _, img := range r.Product.Images {
		if img.ID <= 0 {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_valid_id"}
		}
		if strings.TrimSpace(img.URL) == "" {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_url"}
		}
	}
	return nil
}

func (r *ProductUpdateRequest) validateNewImages() error {
	// Validate each new image
	for _, imageHeader := range r.NewImages {
		// Check file size (max 3MB per image)
		if imageHeader.Size > 3*1024*1024 {
			return &httpErrors.BadRequestError{Message: "image_size_too_large_max_3mb"}
		}

		// Open file to check MIME type
		file, err := imageHeader.Open()
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
		if !isValidImageType(mimeType) {
			return &httpErrors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
		}

		// Reset file pointer for later use
		if seeker, ok := file.(io.Seeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
			}
		}
	}

	return nil
}

// ToImageBuffers converts FileHeaders to byte slices for upload service
func (r *ProductUpdateRequest) ToImageBuffers() ([][]byte, error) {
	buffers := make([][]byte, len(r.NewImages))

	for i, imageHeader := range r.NewImages {
		file, err := imageHeader.Open()
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
		}
		defer file.Close()

		buffer := &bytes.Buffer{}
		if _, err := io.Copy(buffer, file); err != nil {
			return nil, &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
		}
		buffers[i] = buffer.Bytes()
	}

	return buffers, nil
}
