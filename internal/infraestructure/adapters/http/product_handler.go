package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/requests"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Product handler log field constants
const (
	ProductHandlerField           = "product_handler"
	GetByIDFunctionField          = "get_by_id"
	CreateProductFunctionField    = "create"
	UpdateProductFunctionField    = "update"
	DeleteProductFunctionField    = "delete"
	ParseShopIDSubFuncField       = "parse_shop_id"
	ParseProductIDSubFuncField    = "parse_product_id"
	ParsePaginationSubFuncField   = "parse_pagination_params"
	BuildRequestSubFuncField      = "build_request"
	ConvertImagesToBuffersSubFunc = "convert_images_to_buffers"
)

type ProductHandler struct {
	createProduct             ports.CreateProductUseCase
	getAllByShopIDWithFilters ports.GetAllByShopIDWithFiltersUseCase
	getByID                   ports.GetByIDUseCase
	updateProduct             ports.UpdateProductUseCase
	deleteProduct             ports.DeleteProductUseCase
}

func (p *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

	// Parse multipart form (13MB limit - allows 4 images of 3MB each + product data)
	stepStart := time.Now()
	err := r.ParseMultipartForm(13 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "parse_multipart_form",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 1: Multipart form parsed")

	// Create ProductCreateRequest
	stepStart = time.Now()
	request, err := requests.NewProductCreateRequest(r, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": BuildRequestSubFuncField,
			"error":    err.Error(),
		}).Error("Error building product create request")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "build_request",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 2: Request built")

	// Validate request (includes product data and images)
	stepStart = time.Now()
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     CreateProductFunctionField,
			"sub_func":     "request.Validate",
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Product creation validation failed")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "validate_request",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 3: Request validated")

	// Convert images to buffers for upload service
	stepStart = time.Now()
	imageBuffers, err := request.ToImageBuffers()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": ConvertImagesToBuffersSubFunc,
			"error":    err.Error(),
		}).Error("Error converting images to buffers")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "convert_images_to_buffers",
		"image_count": len(imageBuffers),
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 4: Images converted to buffers")

	// Create product via use case
	stepStart = time.Now()
	createdProduct, err := p.createProduct.Execute(ctx, request.ToModel(), imageBuffers, request.ShopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     CreateProductFunctionField,
			"product_name": request.Product.Name,
			"shop_id":      request.ShopID,
			"error":        err.Error(),
		}).Error("Error creating product")
		httpErrors.HandleError(w, err)
		return
	}
	logs.WithFields(map[string]interface{}{
		"operation":   "execute_use_case",
		"duration_ms": time.Since(stepStart).Milliseconds(),
	}).Debug("Step 5: Use case executed")

	logs.WithFields(map[string]interface{}{
		"operation":         "create_product_total",
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Product creation completed successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := responses.CreateProductResponse{
		Product: responses.ProductResponseFromModel(createdProduct),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": CreateProductFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func NewProductHandler(
	createProductUseCase ports.CreateProductUseCase,
	getAllByShopIDWithFiltersUseCase ports.GetAllByShopIDWithFiltersUseCase,
	getByIDUseCase ports.GetByIDUseCase,
	updateProductUseCase ports.UpdateProductUseCase,
	deleteProductUseCase ports.DeleteProductUseCase,
) *ProductHandler {
	return &ProductHandler{
		createProduct:             createProductUseCase,
		getAllByShopIDWithFilters: getAllByShopIDWithFiltersUseCase,
		getByID:                   getByIDUseCase,
		updateProduct:             updateProductUseCase,
		deleteProduct:             deleteProductUseCase,
	}
}

// parseShopID extracts and validates shop_id from URL path.
// Note: Empty path params are handled by Gorilla Mux (returns 404).
func (p *ProductHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": ParseShopIDSubFuncField,
			"shop_id":  shopIDStr,
			"error":    err,
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}

func (p *ProductHandler) GetAllByShopIDWithFilters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse shop_id from URL path (context parameter, not a filter)
	shopID, err := p.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse query parameters into ProductFiltersRequest
	queryParams := r.URL.Query()
	filtersRequest, err := requests.NewProductFiltersRequest(queryParams)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": "get_all_by_shop_id_with_filters",
			"sub_func": "parse_query_params",
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": "get_all_by_shop_id_with_filters",
			"sub_func": "validate_request",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert to domain model (ShopID passed separately)
	filters := filtersRequest.ToProductFilters()

	// Execute use case (business validation happens in use case layer)
	// ShopID is a context parameter, passed separately from filters
	// Filters passed by value (immutable) - use case calls Validated() internally
	// Lightweight query - no variants for real-time search performance
	// totalCount is only returned on first page (cursor empty), nil on subsequent pages
	products, nextCursor, hasMore, totalCount, err := p.getAllByShopIDWithFilters.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        ProductHandlerField,
			"function":    "get_all_by_shop_id_with_filters",
			"shop_id":     shopID,
			"has_search":  filters.Search != nil,
			"has_filters": filters.CategoryID != nil || filters.IsActive != nil,
			"limit":       filters.Limit,
			"last_id":     filters.LastID,
			"error":       err.Error(),
		}).Error("Error retrieving products with filters")
		httpErrors.HandleError(w, err)
		return
	}

	// Build HTTP response (handler constructs response DTO)
	response := responses.NewListProductsResponse(products, nextCursor, hasMore, totalCount)

	// Log successful search
	logs.WithFields(map[string]interface{}{
		"file":         ProductHandlerField,
		"function":     "get_all_by_shop_id_with_filters",
		"shop_id":      shopID,
		"has_search":   filters.Search != nil,
		"result_count": len(products),
		"has_more":     hasMore,
	}).Debug("Products retrieved successfully with filters (lightweight - no variants)")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": "get_all_by_shop_id_with_filters",
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

func (p *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse and validate product_id
	productID, err := p.parseProductID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	product, err := p.getByID.Execute(ctx, productID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   GetByIDFunctionField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Error retrieving product")
		httpErrors.HandleError(w, err)
		return
	}

	// Return product response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := responses.GetProductResponse{
		Product: responses.ProductResponseFromModel(product),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": GetByIDFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseProductID extracts and validates product_id from URL path.
// Note: Empty path params are handled by Gorilla Mux (returns 404).
func (p *ProductHandler) parseProductID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	productIDStr := vars["product_id"]

	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   ParseProductIDSubFuncField,
			"product_id": productIDStr,
			"error":      err,
		}).Error("Invalid product_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_product_id_format"}
	}

	return productID, nil
}

func (p *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

	// Parse and validate product_id
	productID, err := p.parseProductID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(13 << 20) // 13MB limit
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	// Build product update request (different from create)
	request, err := requests.NewProductUpdateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": BuildRequestSubFuncField,
			"error":    err.Error(),
		}).Error("Error building product update request")
		httpErrors.HandleError(w, err)
		return
	}

	// Set product ID from path param (override any ID in JSON)
	request.Product.ID = productID

	// Validate request
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     UpdateProductFunctionField,
			"sub_func":     "request.Validate",
			"product_id":   productID,
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Product update validation failed")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert new images to buffers for upload service
	imageBuffers, err := request.ToImageBuffers()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": ConvertImagesToBuffersSubFunc,
			"error":    err.Error(),
		}).Error("Error converting images to buffers")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
		return
	}

	// Update product via use case
	err = p.updateProduct.Execute(ctx, productID, request.ToModel(), imageBuffers, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductHandlerField,
			"function":     UpdateProductFunctionField,
			"product_id":   productID,
			"product_name": request.Product.Name,
			"error":        err.Error(),
		}).Error("Error updating product")
		httpErrors.HandleError(w, err)
		return
	}

	// Fetch updated product to return in response
	updatedProduct, err := p.getByID.Execute(ctx, productID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   UpdateProductFunctionField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Error retrieving updated product")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.UpdateProductResponse{
		Product: responses.ProductResponseFromModel(updatedProduct),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": UpdateProductFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// Delete handles DELETE /products/{product_id} requests.
func (p *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get shop_id from context (injected by auth middleware from JWT token)
	shopID := claims.GetFirstShopIDFromContext(ctx)
	if shopID == 0 {
		logs.WithFields(map[string]interface{}{
			"file":     ProductHandlerField,
			"function": DeleteProductFunctionField,
			"error":    "shop_id_not_found_in_context",
		}).Error("Shop ID not found in context")
		httpErrors.HandleError(w, &httpErrors.UnauthorizedError{Message: "shop_id_not_found_in_token"})
		return
	}

	// Parse and validate product_id from URL
	productID, err := p.parseProductID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Delete product via use case
	err = p.deleteProduct.Execute(ctx, productID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductHandlerField,
			"function":   DeleteProductFunctionField,
			"product_id": productID,
			"shop_id":    shopID,
			"error":      err.Error(),
		}).Error("Error deleting product")
		httpErrors.HandleError(w, err)
		return
	}

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
