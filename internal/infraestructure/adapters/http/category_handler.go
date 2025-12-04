package http

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Category handler log field constants
const (
	CategoryHandlerField                = "category_handler"
	CreateCategoryFunctionField         = "create"
	GetAllByShopIDCategoryFunctionField = "get_all_by_shop_id_with_filters"
	BuildCategoryRequestSubFunc         = "build_category_request"
	ConvertImageToBufferSubFunc         = "convert_image_to_buffer"
	ParseCategoryShopIDSubFunc          = "parse_shop_id"
	ParseCategoryQueryParamsSubFunc     = "parse_query_params"
)

// CategoryHandler handles HTTP requests for category endpoints.
type CategoryHandler struct {
	createCategory            ports.CreateCategoryUseCase
	getAllByShopIDWithFilters ports.GetAllCategoriesByShopIDWithFiltersUseCase
}

// NewCategoryHandler creates a new CategoryHandler instance.
func NewCategoryHandler(
	createCategoryUseCase ports.CreateCategoryUseCase,
	getAllByShopIDWithFiltersUseCase ports.GetAllCategoriesByShopIDWithFiltersUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		createCategory:            createCategoryUseCase,
		getAllByShopIDWithFilters: getAllByShopIDWithFiltersUseCase,
	}
}

// Create handles POST /categories requests.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form (4MB limit - 1 image of 3MB + category data)
	err := r.ParseMultipartForm(4 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	// Build CategoryCreateRequest
	request, err := h.buildCategoryCreateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"sub_func": BuildCategoryRequestSubFunc,
			"error":    err.Error(),
		}).Error("Error building category create request")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate request (HTTP validation: required fields, image format)
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":          CategoryHandlerField,
			"function":      CreateCategoryFunctionField,
			"sub_func":      "request.Validate",
			"category_name": request.Category.Name,
			"error":         err.Error(),
		}).Error("Category creation validation failed")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert image to buffer for upload service
	imageBuffer, err := request.ToImageBuffer()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"sub_func": ConvertImageToBufferSubFunc,
			"error":    err.Error(),
		}).Error("Error converting image to buffer")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}

	// Create category via use case
	createdCategory, err := h.createCategory.Execute(ctx, &request.Category, imageBuffer, request.ShopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":          CategoryHandlerField,
			"function":      CreateCategoryFunctionField,
			"category_name": request.Category.Name,
			"shop_id":       request.ShopID,
			"error":         err.Error(),
		}).Error("Error creating category")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":        CategoryHandlerField,
		"function":    CreateCategoryFunctionField,
		"category_id": createdCategory.ID,
		"shop_id":     request.ShopID,
	}).Info("Category created successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdCategory); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// buildCategoryCreateRequest builds a CategoryCreateRequest from the HTTP request.
func (h *CategoryHandler) buildCategoryCreateRequest(r *http.Request) (*contracts.CategoryCreateRequest, error) {
	// Extract category JSON from form data
	categoryJSON := r.FormValue("category")
	if strings.TrimSpace(categoryJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "category_json_required"}
	}

	// Parse category JSON
	var category models.Category
	if err := json.Unmarshal([]byte(categoryJSON), &category); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_category_json_format"}
	}

	// Get shop ID from form
	shopIDStr := r.FormValue("shop_id")
	if strings.TrimSpace(shopIDStr) == "" {
		return nil, &httpErrors.BadRequestError{Message: "shop_id_required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
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

	return &contracts.CategoryCreateRequest{
		Category: category,
		ShopID:   shopID,
		Image:    imageHeader,
	}, nil
}

// GetAllByShopIDWithFilters handles GET /shops/{shop_id}/categories requests.
func (h *CategoryHandler) GetAllByShopIDWithFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse shop_id from URL path
	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse query parameters into CategoryFiltersRequest
	filtersRequest, err := contracts.NewCategoryFiltersRequest(r.URL.Query())
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": GetAllByShopIDCategoryFunctionField,
			"sub_func": ParseCategoryQueryParamsSubFunc,
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": GetAllByShopIDCategoryFunctionField,
			"sub_func": "validate_request",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert to domain model
	filters := filtersRequest.ToCategoryFilters()

	// Execute use case
	categories, nextCursor, hasMore, totalCount, err := h.getAllByShopIDWithFilters.Execute(ctx, shopID, &filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       CategoryHandlerField,
			"function":   GetAllByShopIDCategoryFunctionField,
			"shop_id":    shopID,
			"has_search": filters.Search != nil,
			"limit":      filters.Limit,
			"error":      err.Error(),
		}).Error("Error retrieving categories with filters")
		httpErrors.HandleError(w, err)
		return
	}

	// Build HTTP response
	response := contracts.PaginatedCategoriesResponse{
		Categories: categories,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}

	logs.WithFields(map[string]interface{}{
		"file":         CategoryHandlerField,
		"function":     GetAllByShopIDCategoryFunctionField,
		"shop_id":      shopID,
		"has_search":   filters.Search != nil,
		"result_count": len(categories),
		"has_more":     hasMore,
	}).Debug("Categories retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": GetAllByShopIDCategoryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseShopID extracts and validates shop_id from URL path.
func (h *CategoryHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]
	if strings.TrimSpace(shopIDStr) == "" {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": ParseCategoryShopIDSubFunc,
			"error":    "shop_id_parameter_required",
		}).Error("Missing shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "shop_id_parameter_required"}
	}

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": ParseCategoryShopIDSubFunc,
			"sub_func": "strconv.Atoi",
			"shop_id":  shopIDStr,
			"error":    err.Error(),
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}
