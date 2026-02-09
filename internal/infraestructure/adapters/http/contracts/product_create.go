package contracts

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// ProductRequest represents product data in HTTP requests.
type ProductRequest struct {
	ID               int                      `json:"id,omitempty"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	Price            float64                  `json:"price"`
	Category         *ProductCategoryRequest  `json:"category,omitempty"`
	Variants         []*ProductVariantRequest `json:"variants,omitempty"`
	IsActive         bool                     `json:"is_active"`
	IsPromotional    bool                     `json:"is_promotional"`
	PromotionalPrice float64                  `json:"promotional_price,omitempty"`
	IsHighlighted    bool                     `json:"is_highlighted"`
	IsStockeable     bool                     `json:"is_stockeable"`
	Stock            int                      `json:"stock"`
	MinimumStock     int                      `json:"minimum_stock,omitempty"`
	Images           []*ProductImageRequest   `json:"images,omitempty"`
}

// ProductCategoryRequest represents category reference in product requests.
type ProductCategoryRequest struct {
	ID int `json:"id"`
}

// ProductVariantRequest represents variant data in product requests.
type ProductVariantRequest struct {
	ID            int                     `json:"id,omitempty"`
	Name          string                  `json:"name"`
	Order         int                     `json:"order,omitempty"`
	SelectionType string                  `json:"selection_type"`
	MaxSelections int                     `json:"max_selections,omitempty"`
	IsRequired    bool                    `json:"is_required,omitempty"`
	Options       []*ProductOptionRequest `json:"options,omitempty"`
}

// ProductOptionRequest represents option data in variant requests.
type ProductOptionRequest struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name"`
	Price float64 `json:"price,omitempty"`
	Order int     `json:"order,omitempty"`
}

// ProductImageRequest represents image data in product requests.
type ProductImageRequest struct {
	ID    int    `json:"id,omitempty"`
	URL   string `json:"url,omitempty"`
	Order int    `json:"order,omitempty"`
}

// ToModel converts ProductRequest to a domain model.
func (p *ProductRequest) ToModel() *models.Product {
	product := &models.Product{
		ID:               p.ID,
		Name:             p.Name,
		Description:      p.Description,
		Price:            p.Price,
		IsActive:         p.IsActive,
		IsPromotional:    p.IsPromotional,
		PromotionalPrice: p.PromotionalPrice,
		IsHighlighted:    p.IsHighlighted,
		IsStockeable:     p.IsStockeable,
		Stock:            p.Stock,
		MinimumStock:     p.MinimumStock,
	}

	if p.Category != nil {
		product.Category = &models.Category{ID: p.Category.ID}
	}

	// Convert variants
	if len(p.Variants) > 0 {
		product.Variants = make([]*models.Variant, len(p.Variants))
		for i, v := range p.Variants {
			product.Variants[i] = v.ToModel()
		}
	}

	// Convert images
	if len(p.Images) > 0 {
		product.Images = make([]*models.Image, len(p.Images))
		for i, img := range p.Images {
			product.Images[i] = &models.Image{
				ID:  img.ID,
				URL: img.URL,
			}
		}
	}

	return product
}

// ToModel converts ProductVariantRequest to a domain model.
func (v *ProductVariantRequest) ToModel() *models.Variant {
	variant := &models.Variant{
		ID:            v.ID,
		Name:          v.Name,
		Order:         v.Order,
		SelectionType: models.SelectionType(v.SelectionType),
		MaxSelections: v.MaxSelections,
		IsRequired:    v.IsRequired,
	}

	if len(v.Options) > 0 {
		variant.Options = make([]*models.Option, len(v.Options))
		for i, opt := range v.Options {
			variant.Options[i] = &models.Option{
				ID:    opt.ID,
				Name:  opt.Name,
				Price: opt.Price,
				Order: opt.Order,
			}
		}
	}

	return variant
}

type ProductCreateRequest struct {
	Product ProductRequest          `json:"product"`
	ShopID  int                     `json:"shop_id"`
	Images  []*multipart.FileHeader `json:"-"` // Not part of JSON, set manually
}

// NewProductCreateRequest creates a ProductCreateRequest from an HTTP request.
// shopID comes from the JWT token context (injected by auth middleware).
func NewProductCreateRequest(r *http.Request, shopID int) (*ProductCreateRequest, error) {
	// Parse product JSON from form data
	productJSON := r.FormValue("product")
	if strings.TrimSpace(productJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "product_json_required"}
	}

	var product ProductRequest
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_product_json_format"}
	}

	// Parse images from form
	var images []*multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for i := 0; ; i++ {
			key := "images[" + strconv.Itoa(i) + "]"
			files, exists := r.MultipartForm.File[key]
			if !exists || len(files) == 0 {
				break
			}
			images = append(images, files[0])
		}
	}
	if len(images) == 0 {
		return nil, &httpErrors.BadRequestError{Message: "product_image_required"}
	}

	return &ProductCreateRequest{
		Product: product,
		ShopID:  shopID,
		Images:  images,
	}, nil
}

// ToModel converts the request's product to a domain model.
func (r *ProductCreateRequest) ToModel() *models.Product {
	return r.Product.ToModel()
}

func (r *ProductCreateRequest) Validate() error {
	// Validate product data
	if err := r.validateProduct(); err != nil {
		return err
	}

	// Validate shop ID
	if r.ShopID <= 0 {
		return &httpErrors.BadRequestError{Message: "shop_id_is_required"}
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

	if err := r.validateVariants(); err != nil {
		return err
	}

	// Note: Product.Validate() will be called in the service layer
	// to validate business rules (price, stock, promotional price, etc.)
	return nil
}

func (r *ProductCreateRequest) validateBasicProductFields() error {
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

func (r *ProductCreateRequest) validateVariants() error {
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

func isValidSelectionType(selectionType string) bool {
	switch selectionType {
	case "single", "multiple", "custom":
		return true
	default:
		return false
	}
}

func (r *ProductCreateRequest) validateVariantOptions(variant *ProductVariantRequest, variantIndex int) error {
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

func (r *ProductCreateRequest) validateImages() error {
	if len(r.Images) == 0 {
		return &httpErrors.BadRequestError{Message: "product_image_required"}
	}

	// Validate each image
	for _, imageHeader := range r.Images {
		// Check file size (max 5MB per image)
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
