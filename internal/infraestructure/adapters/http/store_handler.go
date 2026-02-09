package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Store handler log field constants
const (
	StoreHandlerField                     = "store_handler"
	GetStoreBySlugFunctionField           = "get_by_slug"
	GetStoreCategoriesFunctionField       = "get_categories"
	GetStoreProductsFunctionField         = "get_products"
	GetStoreFeaturedProductsFunctionField = "get_featured_products"
	GetStoreProductByIDFunctionField      = "get_product_by_id"
)

type StoreHandler struct {
	getStoreBySlug           ports.GetStoreBySlugUseCase
	getStoreCategories       ports.GetStoreCategoriesUseCase
	getStoreProducts         ports.GetStoreProductsUseCase
	getStoreFeaturedProducts ports.GetStoreFeaturedProductsUseCase
	getStoreProductByID      ports.GetStoreProductByIDUseCase
}

func NewStoreHandler(
	getStoreBySlug ports.GetStoreBySlugUseCase,
	getStoreCategories ports.GetStoreCategoriesUseCase,
	getStoreProducts ports.GetStoreProductsUseCase,
	getStoreFeaturedProducts ports.GetStoreFeaturedProductsUseCase,
	getStoreProductByID ports.GetStoreProductByIDUseCase,
) ports.StoreHandler {
	return &StoreHandler{
		getStoreBySlug:           getStoreBySlug,
		getStoreCategories:       getStoreCategories,
		getStoreProducts:         getStoreProducts,
		getStoreFeaturedProducts: getStoreFeaturedProducts,
		getStoreProductByID:      getStoreProductByID,
	}
}

func (h *StoreHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case (NO authorization check - public endpoint)
	store, err := h.getStoreBySlug.Execute(ctx, slug)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreBySlugFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Error retrieving store")
		httpErrors.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(store)
}

func (h *StoreHandler) parseSlug(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	// Validate slug is not empty
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", &httpErrors.BadRequestError{Message: "invalid_slug_format"}
	}

	return slug, nil
}

// GetCategories handles GET /stores/{slug}/categories requests.
func (h *StoreHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse query parameters into CategoryFiltersRequest
	filtersRequest, err := contracts.NewCategoryFiltersRequest(r.URL.Query())
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreCategoriesFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreCategoriesFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert to domain model
	filters := filtersRequest.ToCategoryFilters()

	// Execute use case (NO authorization check - public endpoint)
	categories, nextCursor, hasMore, err := h.getStoreCategories.Execute(ctx, slug, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       StoreHandlerField,
			"function":   GetStoreCategoriesFunctionField,
			"slug":       slug,
			"has_search": filters.Search != nil,
			"limit":      filters.Limit,
			"error":      err.Error(),
		}).Error("Error retrieving store categories")
		httpErrors.HandleError(w, err)
		return
	}

	// Build HTTP response (no TotalCount - store frontend doesn't use pagination UI)
	response := contracts.NewPaginatedCategoriesResponse(categories, nextCursor, hasMore, nil)

	logs.WithFields(map[string]interface{}{
		"file":         StoreHandlerField,
		"function":     GetStoreCategoriesFunctionField,
		"slug":         slug,
		"has_search":   filters.Search != nil,
		"result_count": len(categories),
		"has_more":     hasMore,
	}).Debug("Store categories retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreCategoriesFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetProducts handles GET /stores/{slug}/products requests.
// Returns all products for a store with optional filters.
func (h *StoreHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse query parameters into ProductFiltersRequest
	filtersRequest, err := contracts.NewProductFiltersRequest(r.URL.Query())
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreProductsFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreProductsFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert to domain model
	filters := filtersRequest.ToProductFilters()

	// Execute use case (NO authorization check - public endpoint)
	products, nextCursor, hasMore, err := h.getStoreProducts.Execute(ctx, slug, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreProductsFunctionField,
			"slug":     slug,
			"limit":    filters.Limit,
			"error":    err.Error(),
		}).Error("Error retrieving store products")
		httpErrors.HandleError(w, err)
		return
	}

	// Build HTTP response (no TotalCount - store frontend doesn't need it)
	response := contracts.NewPaginatedProductsResponse(products, nextCursor, hasMore, nil)

	logs.WithFields(map[string]interface{}{
		"file":         StoreHandlerField,
		"function":     GetStoreProductsFunctionField,
		"slug":         slug,
		"result_count": len(response.Products),
		"has_more":     hasMore,
	}).Debug("Store products retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreProductsFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetFeaturedProducts handles GET /stores/{slug}/products/featured requests.
// Returns products where is_highlighted=true and is_active=true.
func (h *StoreHandler) GetFeaturedProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse query parameters into ProductFiltersRequest
	filtersRequest, err := contracts.NewProductFiltersRequest(r.URL.Query())
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreFeaturedProductsFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreFeaturedProductsFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert to domain model (is_highlighted and is_active are forced in use case)
	filters := filtersRequest.ToProductFilters()

	// Execute use case (NO authorization check - public endpoint)
	products, nextCursor, hasMore, err := h.getStoreFeaturedProducts.Execute(ctx, slug, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreFeaturedProductsFunctionField,
			"slug":     slug,
			"limit":    filters.Limit,
			"error":    err.Error(),
		}).Error("Error retrieving store featured products")
		httpErrors.HandleError(w, err)
		return
	}

	// Build HTTP response (no TotalCount - store frontend doesn't need it)
	response := contracts.NewPaginatedProductsResponse(products, nextCursor, hasMore, nil)

	logs.WithFields(map[string]interface{}{
		"file":         StoreHandlerField,
		"function":     GetStoreFeaturedProductsFunctionField,
		"slug":         slug,
		"result_count": len(response.Products),
		"has_more":     hasMore,
	}).Debug("Store featured products retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreFeaturedProductsFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetProductByID handles GET /stores/{slug}/products/{productId} requests.
// Returns a product by ID for a store identified by slug.
func (h *StoreHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse productId from URL
	vars := mux.Vars(r)
	productIDStr := vars["productId"]
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":       StoreHandlerField,
			"function":   GetStoreProductByIDFunctionField,
			"slug":       slug,
			"product_id": productIDStr,
		}).Warn("Invalid product ID format")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_product_id"})
		return
	}

	// Execute use case (NO authorization check - public endpoint)
	product, err := h.getStoreProductByID.Execute(ctx, slug, productID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       StoreHandlerField,
			"function":   GetStoreProductByIDFunctionField,
			"slug":       slug,
			"product_id": productID,
			"error":      err.Error(),
		}).Error("Error retrieving store product")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":       StoreHandlerField,
		"function":   GetStoreProductByIDFunctionField,
		"slug":       slug,
		"product_id": productID,
	}).Debug("Store product retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(contracts.ProductResponseFromModel(product)); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StoreHandlerField,
			"function": GetStoreProductByIDFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}
