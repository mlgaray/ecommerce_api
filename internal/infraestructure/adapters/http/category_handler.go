package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/requests"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Category handler log field constants
const (
	CategoryHandlerField                = "category_handler"
	CreateCategoryFunctionField         = "create"
	UpdateCategoryFunctionField         = "update"
	DeleteCategoryFunctionField         = "delete"
	GetByIDCategoryFunctionField        = "get_by_id"
	GetAllByShopIDCategoryFunctionField = "get_all_by_shop_id_with_filters"
	BuildCategoryRequestSubFunc         = "build_category_request"
	BuildCategoryUpdateRequestSubFunc   = "build_category_update_request"
	ConvertImageToBufferSubFunc         = "convert_image_to_buffer"
	ParseCategoryShopIDSubFunc          = "parse_shop_id"
	ParseCategoryIDSubFunc              = "parse_category_id"
	ParseCategoryQueryParamsSubFunc     = "parse_query_params"
)

// CategoryHandler handles HTTP requests for category endpoints.
type CategoryHandler struct {
	createCategory            ports.CreateCategoryUseCase
	updateCategory            ports.UpdateCategoryUseCase
	deleteCategory            ports.DeleteCategoryUseCase
	getByID                   ports.GetCategoryByIDUseCase
	getAllByShopIDWithFilters ports.GetAllCategoriesByShopIDWithFiltersUseCase
}

// NewCategoryHandler creates a new CategoryHandler instance.
func NewCategoryHandler(
	createCategoryUseCase ports.CreateCategoryUseCase,
	updateCategoryUseCase ports.UpdateCategoryUseCase,
	deleteCategoryUseCase ports.DeleteCategoryUseCase,
	getByIDUseCase ports.GetCategoryByIDUseCase,
	getAllByShopIDWithFiltersUseCase ports.GetAllCategoriesByShopIDWithFiltersUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		createCategory:            createCategoryUseCase,
		updateCategory:            updateCategoryUseCase,
		deleteCategory:            deleteCategoryUseCase,
		getByID:                   getByIDUseCase,
		getAllByShopIDWithFilters: getAllByShopIDWithFiltersUseCase,
	}
}

// Create handles POST /categories requests.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

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
	request, err := requests.NewCategoryCreateRequest(r, shopID)
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
	createdCategory, err := h.createCategory.Execute(ctx, request.ToModel(), imageBuffer, request.ShopID)
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
	response := responses.CreateCategoryResponse{
		Category: responses.CategoryResponseFromModel(createdCategory),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": CreateCategoryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
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
	filtersRequest, err := requests.NewCategoryFiltersRequest(r.URL.Query())
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

	// Execute use case (filters passed by value - immutable)
	categories, nextCursor, hasMore, totalCount, err := h.getAllByShopIDWithFilters.Execute(ctx, shopID, filters)
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
	response := responses.NewListCategoriesResponse(categories, nextCursor, hasMore, totalCount)

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

// GetByID handles GET /categories/{category_id} requests.
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate category_id from URL
	categoryID, err := h.parseCategoryID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	category, err := h.getByID.Execute(ctx, categoryID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryHandlerField,
			"function":    GetByIDCategoryFunctionField,
			"category_id": categoryID,
			"error":       err.Error(),
		}).Error("Error retrieving category")
		httpErrors.HandleError(w, err)
		return
	}

	// Return category response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := responses.GetCategoryResponse{
		Category: responses.CategoryResponseFromModel(category),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": GetByIDCategoryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// Update handles PUT /categories/{category_id} requests.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": UpdateCategoryFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

	// Parse and validate category_id from URL
	categoryID, err := h.parseCategoryID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse multipart form (4MB limit - 1 image of 3MB + category data)
	err = r.ParseMultipartForm(4 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": UpdateCategoryFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	// Build CategoryUpdateRequest
	request, err := requests.NewCategoryUpdateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": UpdateCategoryFunctionField,
			"sub_func": BuildCategoryUpdateRequestSubFunc,
			"error":    err.Error(),
		}).Error("Error building category update request")
		httpErrors.HandleError(w, err)
		return
	}

	// Set category ID from path param (override any ID in JSON)
	request.Category.ID = categoryID

	// Validate request (HTTP validation: required fields, image format)
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":          CategoryHandlerField,
			"function":      UpdateCategoryFunctionField,
			"sub_func":      "request.Validate",
			"category_id":   categoryID,
			"category_name": request.Category.Name,
			"error":         err.Error(),
		}).Error("Category update validation failed")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert new image to buffer for upload service (optional)
	imageBuffer, err := request.ToImageBuffer()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": UpdateCategoryFunctionField,
			"sub_func": ConvertImageToBufferSubFunc,
			"error":    err.Error(),
		}).Error("Error converting image to buffer")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}

	// Update category via use case
	err = h.updateCategory.Execute(ctx, categoryID, request.ToModel(), imageBuffer, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":          CategoryHandlerField,
			"function":      UpdateCategoryFunctionField,
			"category_id":   categoryID,
			"category_name": request.Category.Name,
			"shop_id":       shopID,
			"error":         err.Error(),
		}).Error("Error updating category")
		httpErrors.HandleError(w, err)
		return
	}

	updatedCategory, err := h.getByID.Execute(ctx, categoryID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryHandlerField,
			"function":    UpdateCategoryFunctionField,
			"category_id": categoryID,
			"shop_id":     shopID,
			"error":       err.Error(),
		}).Error("Error retrieving updated category")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":        CategoryHandlerField,
		"function":    UpdateCategoryFunctionField,
		"category_id": categoryID,
		"shop_id":     shopID,
	}).Info("Category updated successfully")

	response := responses.UpdateCategoryResponse{
		Category: responses.CategoryResponseFromModel(updatedCategory),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": UpdateCategoryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseCategoryID extracts and validates category_id from URL path.
// Note: Empty path params are handled by Gorilla Mux (returns 404).
func (h *CategoryHandler) parseCategoryID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	categoryIDStr := vars["category_id"]

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryHandlerField,
			"function":    ParseCategoryIDSubFunc,
			"category_id": categoryIDStr,
			"error":       err,
		}).Error("Invalid category_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_category_id_format"}
	}

	return categoryID, nil
}

// parseShopID extracts and validates shop_id from URL path.
// Note: Empty path params are handled by Gorilla Mux (returns 404).
func (h *CategoryHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": ParseCategoryShopIDSubFunc,
			"shop_id":  shopIDStr,
			"error":    err,
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}

// Delete handles DELETE /categories/{category_id} requests.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     CategoryHandlerField,
			"function": DeleteCategoryFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

	// Parse and validate category_id from URL
	categoryID, err := h.parseCategoryID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Delete category via use case
	err = h.deleteCategory.Execute(ctx, categoryID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        CategoryHandlerField,
			"function":    DeleteCategoryFunctionField,
			"category_id": categoryID,
			"shop_id":     shopID,
			"error":       err.Error(),
		}).Error("Error deleting category")
		httpErrors.HandleError(w, err)
		return
	}

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
