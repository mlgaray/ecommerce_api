package contracts

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductCreateRequest struct {
	Product models.Product          `json:"product"`
	ShopID  int                     `json:"shop_id"`
	Images  []*multipart.FileHeader `json:"-"` // Not part of JSON, set manually
}

func (r *ProductCreateRequest) Validate() error {
	// Validate product data
	if err := r.validateProduct(); err != nil {
		return err
	}

	// Validate shop ID
	if r.ShopID <= 0 {
		return &errors.BadRequestError{Message: "shop_id_is_required"}
	}

	// Validate images
	if err := r.validateImages(); err != nil {
		return err
	}

	return nil
}

func (r *ProductCreateRequest) validateProduct() error {
	if err := r.validateBasicProductFields(); err != nil {
		return err
	}

	if err := r.validatePromotionalPrice(); err != nil {
		return err
	}

	if err := r.validateVariants(); err != nil {
		return err
	}

	return nil
}

func (r *ProductCreateRequest) validateBasicProductFields() error {
	if strings.TrimSpace(r.Product.Name) == "" {
		return &errors.BadRequestError{Message: "product_name_is_required"}
	}
	if strings.TrimSpace(r.Product.Description) == "" {
		return &errors.BadRequestError{Message: "product_description_is_required"}
	}
	if r.Product.Price <= 0 {
		return &errors.BadRequestError{Message: "product_price_must_be_positive"}
	}
	if r.Product.Stock < 0 {
		return &errors.BadRequestError{Message: "product_stock_cannot_be_negative"}
	}
	if r.Product.MinimumStock < 0 {
		return &errors.BadRequestError{Message: "product_minimum_stock_cannot_be_negative"}
	}
	if r.Product.Stock > 0 && r.Product.MinimumStock > r.Product.Stock {
		return &errors.BadRequestError{Message: "product_minimum_stock_cannot_be_greater_than_stock"}
	}
	if r.Product.Category == nil || r.Product.Category.ID <= 0 {
		return &errors.BadRequestError{Message: "category_id_is_required"}
	}
	return nil
}

func (r *ProductCreateRequest) validatePromotionalPrice() error {
	if r.Product.IsPromotional && r.Product.PromotionalPrice <= 0 {
		return &errors.BadRequestError{Message: "promotional_price_must_be_positive_when_promotional"}
	}
	return nil
}

func (r *ProductCreateRequest) validateVariants() error {
	for i, variant := range r.Product.Variants {
		if strings.TrimSpace(variant.Name) == "" {
			return &errors.BadRequestError{Message: "variant_name_is_required"}
		}
		if variant.SelectionType == "" {
			return &errors.BadRequestError{Message: "variant_selection_type_is_required"}
		}
		if len(variant.Options) == 0 {
			return &errors.BadRequestError{Message: "variant_must_have_at_least_one_option"}
		}

		if err := r.validateVariantOptions(variant, i); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductCreateRequest) validateVariantOptions(variant *models.Variant, variantIndex int) error {
	for j, option := range variant.Options {
		if strings.TrimSpace(option.Name) == "" {
			return &errors.BadRequestError{Message: "option_name_is_required"}
		}
		if option.Price < 0 {
			return &errors.BadRequestError{Message: "option_price_cannot_be_negative"}
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

func (r *ProductCreateRequest) validateImages() error {
	if len(r.Images) == 0 {
		return &errors.BadRequestError{Message: "at_least_one_image_is_required"}
	}

	// Validate each image
	for _, imageHeader := range r.Images {
		// Check file size (max 5MB per image)
		if imageHeader.Size > 3*1024*1024 {
			return &errors.BadRequestError{Message: "image_size_too_large_max_3mb"}
		}

		// Open file to check MIME type
		file, err := imageHeader.Open()
		if err != nil {
			return &errors.BadRequestError{Message: "cannot_open_image_file"}
		}
		defer file.Close()

		// Read first 512 bytes to detect MIME type
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			return &errors.BadRequestError{Message: "cannot_read_image_file"}
		}

		// Check MIME type
		mimeType := http.DetectContentType(buffer)
		if !isValidImageType(mimeType) {
			return &errors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
		}

		// Reset file pointer for later use
		if seeker, ok := file.(io.Seeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return &errors.BadRequestError{Message: "cannot_reset_file_pointer"}
			}
		}
	}

	return nil
}

func isValidImageType(mimeType string) bool {
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

// ToImageBuffers converts FileHeaders to byte slices for upload service
func (r *ProductCreateRequest) ToImageBuffers() ([][]byte, error) {
	buffers := make([][]byte, len(r.Images))

	for i, imageHeader := range r.Images {
		file, err := imageHeader.Open()
		if err != nil {
			return nil, &errors.BadRequestError{Message: "cannot_open_image_file"}
		}
		defer file.Close()

		buffer := &bytes.Buffer{}
		if _, err := io.Copy(buffer, file); err != nil {
			return nil, &errors.BadRequestError{Message: "cannot_read_image_file"}
		}
		buffers[i] = buffer.Bytes()
	}

	return buffers, nil
}
