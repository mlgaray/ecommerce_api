package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Order handler log field constants
const (
	OrderHandlerField        = "order_handler"
	CreateOrderFunctionField = "create"
)

type OrderHandler struct {
	createOrderUseCase ports.CreateOrderUseCase
}

func NewOrderHandler(createOrderUseCase ports.CreateOrderUseCase) ports.OrderHandler {
	return &OrderHandler{
		createOrderUseCase: createOrderUseCase,
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
