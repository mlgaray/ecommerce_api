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

func isValidSelectionType(selectionType string) bool {
	switch selectionType {
	case "single", "multiple", "custom":
		return true
	default:
		return false
	}
}

// =============================================================================
// Create
// =============================================================================

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

// =============================================================================
// Filters
// =============================================================================

// ProductFiltersRequest represents HTTP query parameters for product search/filtering
// This validates HTTP-specific concerns (format, types), NOT business rules
// Note: ShopID is NOT included here - it's a context parameter passed separately
// from the URL path (e.g., /shops/{shop_id}/products)
type ProductFiltersRequest struct {
	Search        *string  // Optional search term
	CategoryID    *int     // Optional category filter
	IsActive      *bool    // Optional active status filter
	IsHighlighted *bool    // Optional highlighted filter
	IsPromotional *bool    // Optional promotional filter
	MinPrice      *float64 // Optional minimum price
	MaxPrice      *float64 // Optional maximum price
	SortBy        string   // Sort field: "price", "name", "created_at"
	SortOrder     string   // Sort order: "asc", "desc"
	Limit         int      // Items per page (default: 20)
	Cursor        string   // Pagination cursor (opaque base64 string)
}

// Validate validates HTTP-specific concerns
// HTTP validations: format, types, required HTTP fields
// Business validations happen in models.ProductFilters.Validate()
// Note: ShopID validation is NOT here - it's validated separately in the handler
func (r *ProductFiltersRequest) Validate() error {
	// HTTP validation: if search is provided, it should not be empty or only whitespace
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		// Update to trimmed value
		*r.Search = trimmed
	}

	// HTTP validation: limit should be a valid integer if provided
	// Actual business rules (min/max) are handled in domain layer
	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	// HTTP validation: cursor is an opaque string, no format validation needed here
	// Decoding happens in ToProductFilters() method - invalid cursors are treated as first page

	// HTTP validation: prices should be valid numbers if provided
	if r.MinPrice != nil && *r.MinPrice < 0 {
		return &httpErrors.BadRequestError{Message: "min_price_cannot_be_negative"}
	}

	if r.MaxPrice != nil && *r.MaxPrice < 0 {
		return &httpErrors.BadRequestError{Message: "max_price_cannot_be_negative"}
	}

	// All other business rules (price range logic, sort field validation, etc.)
	// are handled in models.ProductFilters.Validate()

	return nil
}

// ToProductFilters converts HTTP request to domain model
// Decodes the opaque cursor (if present) into LastID and LastSortValue
// Note: ShopID is NOT included - it's passed separately as a context parameter
func (r *ProductFiltersRequest) ToProductFilters() models.ProductFilters {
	filters := models.ProductFilters{
		Search:        r.Search,
		CategoryID:    r.CategoryID,
		IsActive:      r.IsActive,
		IsHighlighted: r.IsHighlighted,
		IsPromotional: r.IsPromotional,
		MinPrice:      r.MinPrice,
		MaxPrice:      r.MaxPrice,
		SortBy:        r.SortBy,
		SortOrder:     r.SortOrder,
		Limit:         r.Limit,
		LastID:        nil, // Will be populated from cursor
		LastSortValue: nil, // Will be populated from cursor
	}

	// Decode cursor if present
	// HTTP layer responsibility: convert opaque cursor string to primitive values
	if r.Cursor != "" {
		cursorData, err := pagination.DecodeCursor(r.Cursor)
		if err != nil {
			// If cursor is invalid, treat as first page (LastID = nil)
			// This is graceful degradation - don't fail the request
			return filters
		}

		// Populate decoded cursor values
		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}

// NewProductFiltersRequest parses query parameters from URL into ProductFiltersRequest
// Note: ShopID is NOT included here - it's a context parameter parsed separately from URL path
func NewProductFiltersRequest(queryParams map[string][]string) (*ProductFiltersRequest, error) {
	request := &ProductFiltersRequest{}

	request.parseSearchParam(queryParams)

	if err := request.parseFilterParams(queryParams); err != nil {
		return nil, err
	}

	if err := request.parsePriceParams(queryParams); err != nil {
		return nil, err
	}

	if err := request.parsePaginationParams(queryParams); err != nil {
		return nil, err
	}

	return request, nil
}

// parseSearchParam extracts search query parameter
func (r *ProductFiltersRequest) parseSearchParam(queryParams map[string][]string) {
	if search := getQueryParam(queryParams, "search"); search != "" {
		r.Search = &search
	}
}

// parseFilterParams extracts category and boolean filter parameters
func (r *ProductFiltersRequest) parseFilterParams(queryParams map[string][]string) error {
	// Parse category_id
	if categoryIDStr := getQueryParam(queryParams, "category_id"); categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_category_id_format"}
		}
		r.CategoryID = &categoryID
	}

	// Parse boolean filters
	if err := r.parseBoolParam(queryParams, "is_active", &r.IsActive); err != nil {
		return err
	}
	if err := r.parseBoolParam(queryParams, "is_highlighted", &r.IsHighlighted); err != nil {
		return err
	}
	if err := r.parseBoolParam(queryParams, "is_promotional", &r.IsPromotional); err != nil {
		return err
	}

	return nil
}

// parseBoolParam parses a boolean query parameter into the target pointer
func (r *ProductFiltersRequest) parseBoolParam(queryParams map[string][]string, key string, target **bool) error {
	if str := getQueryParam(queryParams, key); str != "" {
		value, err := strconv.ParseBool(str)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_" + key + "_format"}
		}
		*target = &value
	}
	return nil
}

// parsePriceParams extracts price range parameters
func (r *ProductFiltersRequest) parsePriceParams(queryParams map[string][]string) error {
	if minPriceStr := getQueryParam(queryParams, "min_price"); minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_min_price_format"}
		}
		r.MinPrice = &minPrice
	}

	if maxPriceStr := getQueryParam(queryParams, "max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_max_price_format"}
		}
		r.MaxPrice = &maxPrice
	}

	return nil
}

// parsePaginationParams extracts sorting and pagination parameters
func (r *ProductFiltersRequest) parsePaginationParams(queryParams map[string][]string) error {
	r.SortBy = getQueryParam(queryParams, "sort")
	r.SortOrder = getQueryParam(queryParams, "order")

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		r.Limit = limit
	}

	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		r.Cursor = cursorStr
	}

	return nil
}

// getQueryParam safely retrieves a single query parameter value
func getQueryParam(queryParams map[string][]string, key string) string {
	values, exists := queryParams[key]
	if !exists || len(values) == 0 {
		return ""
	}
	return values[0]
}

// =============================================================================
// Update
// =============================================================================

type ProductUpdateRequest struct {
	Product   ProductRequest          `json:"product"`
	NewImages []*multipart.FileHeader `json:"-"` // Optional new images to upload
}

// NewProductUpdateRequest creates a ProductUpdateRequest from an HTTP request.
func NewProductUpdateRequest(r *http.Request) (*ProductUpdateRequest, error) {
	// Extract product JSON from form data
	productJSON := r.FormValue("product")
	if strings.TrimSpace(productJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "product_json_required"}
	}

	// Parse product JSON (includes existing images with IDs)
	var product ProductRequest
	if err := json.Unmarshal([]byte(productJSON), &product); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_product_json_format"}
	}

	// Get new images from multipart form (optional for update)
	var newImages []*multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for i := 0; ; i++ {
			key := "images[" + strconv.Itoa(i) + "]"
			if files, exists := r.MultipartForm.File[key]; exists && len(files) > 0 {
				newImages = append(newImages, files[0])
			} else {
				break
			}
		}
	}

	return &ProductUpdateRequest{
		Product:   product,
		NewImages: newImages,
	}, nil
}

// ToModel converts the request's product to a domain model.
func (r *ProductUpdateRequest) ToModel() *models.Product {
	return r.Product.ToModel()
}

func (r *ProductUpdateRequest) Validate() error {
	// Validate product data
	if err := r.validateProduct(); err != nil {
		return err
	}

	// CRITICAL: Validate that at least one image exists (existing OR new)
	// User could delete all existing images, so we need at least one new image
	if len(r.Product.Images) == 0 && len(r.NewImages) == 0 {
		return &httpErrors.BadRequestError{Message: "product_image_required"}
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

func (r *ProductUpdateRequest) validateVariantOptions(variant *ProductVariantRequest, variantIndex int) error {
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
