package http

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Category handler log field constants
const (
	CategoryHandlerField        = "category_handler"
	CreateCategoryFunctionField = "create"
	BuildCategoryRequestSubFunc = "build_category_request"
	ConvertImageToBufferSubFunc = "convert_image_to_buffer"
)

// CategoryHandler handles HTTP requests for category endpoints.
type CategoryHandler struct {
	createCategory ports.CreateCategoryUseCase
}

// NewCategoryHandler creates a new CategoryHandler instance.
func NewCategoryHandler(createCategoryUseCase ports.CreateCategoryUseCase) *CategoryHandler {
	return &CategoryHandler{
		createCategory: createCategoryUseCase,
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
