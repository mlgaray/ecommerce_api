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

// Order handler log field constants
const (
	OrderHandlerField            = "order_handler"
	CreateOrderFunctionField     = "create"
	GetAllOrdersFunctionField    = "get_all"
	ParseOrderShopIDSubFuncField = "parse_shop_id"
)

type OrderHandler struct {
	createOrderUseCase    ports.CreateOrderUseCase
	getAllByShopIDUseCase ports.GetAllOrdersByShopIDUseCase
}

func NewOrderHandler(
	createOrderUseCase ports.CreateOrderUseCase,
	getAllByShopIDUseCase ports.GetAllOrdersByShopIDUseCase,
) ports.OrderHandler {
	return &OrderHandler{
		createOrderUseCase:    createOrderUseCase,
		getAllByShopIDUseCase: getAllByShopIDUseCase,
	}
}

// Create handles POST /stores/{slug}/orders requests.
// Public endpoint - no authentication required.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parse store slug from URL
	slug, err := h.parseSlug(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// 2. Parse JSON body
	var request contracts.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": CreateOrderFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Warn("Invalid JSON body")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_json_body"})
		return
	}

	// 3. Validate HTTP request (format, types)
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": CreateOrderFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Warn("Invalid request parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// 4. Convert to domain model
	order := request.ToModel()

	// 5. Execute use case
	createdOrder, err := h.createOrderUseCase.Execute(ctx, order, slug)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": CreateOrderFunctionField,
			"slug":     slug,
			"error":    err.Error(),
		}).Error("Error creating order")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":         OrderHandlerField,
		"function":     CreateOrderFunctionField,
		"slug":         slug,
		"order_id":     createdOrder.ID,
		"order_number": createdOrder.OrderNumber,
		"items_count":  len(createdOrder.Items),
		"total":        createdOrder.Total,
	}).Info("Order created successfully")

	response := contracts.OrderResponseFromModel(createdOrder)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": CreateOrderFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseSlug extracts and validates the store slug from the URL.
func (h *OrderHandler) parseSlug(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", &httpErrors.BadRequestError{Message: "invalid_slug_format"}
	}

	return slug, nil
}

// GetAll handles GET /shops/{shop_id}/orders requests.
// Protected endpoint - authentication required.
func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Parse shop_id from URL path
	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// 2. Parse query parameters into OrderFiltersRequest
	queryParams := r.URL.Query()
	filtersRequest, err := contracts.NewOrderFiltersRequest(queryParams)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": GetAllOrdersFunctionField,
			"sub_func": "parse_query_params",
			"error":    err.Error(),
		}).Error("Error parsing query parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// 3. Validate HTTP request (format, types)
	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": GetAllOrdersFunctionField,
			"sub_func": "validate_request",
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	// 4. Convert to domain model (ShopID passed separately)
	filters := filtersRequest.ToOrderFilters()

	// 5. Execute use case (business validation happens in use case layer)
	// ShopID is a context parameter, passed separately from filters
	// Filters passed by value (immutable) - use case calls Validated() internally
	// Lightweight query - no items for dashboard performance
	// totalCount is only returned on first page (cursor empty), nil on subsequent pages
	orders, nextCursor, hasMore, totalCount, err := h.getAllByShopIDUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        OrderHandlerField,
			"function":    GetAllOrdersFunctionField,
			"shop_id":     shopID,
			"has_search":  filters.Search != nil,
			"has_filters": filters.Status != nil || filters.DateFrom != nil,
			"limit":       filters.Limit,
			"last_id":     filters.LastID,
			"error":       err.Error(),
		}).Error("Error retrieving orders with filters")
		httpErrors.HandleError(w, err)
		return
	}

	// 6. Build HTTP response (handler constructs response DTO)
	response := contracts.NewPaginatedOrdersResponse(orders, nextCursor, hasMore, totalCount)

	// Log successful retrieval
	logs.WithFields(map[string]interface{}{
		"file":         OrderHandlerField,
		"function":     GetAllOrdersFunctionField,
		"shop_id":      shopID,
		"has_search":   filters.Search != nil,
		"result_count": len(orders),
		"has_more":     hasMore,
	}).Debug("Orders retrieved successfully with filters (lightweight - no items)")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": GetAllOrdersFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseShopID extracts and validates shop_id from URL path.
// Note: Empty path params are handled by Gorilla Mux (returns 404).
func (h *OrderHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":     OrderHandlerField,
			"function": ParseOrderShopIDSubFuncField,
			"shop_id":  shopIDStr,
			"error":    err,
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}
