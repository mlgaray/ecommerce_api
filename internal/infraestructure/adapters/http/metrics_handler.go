package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/requests"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Metrics handler log field constants
const (
	MetricsHandlerField             = "metrics_handler"
	GetDashboardFunctionField       = "get_dashboard"
	GetRevenueTrendFunctionField    = "get_revenue_trend"
	GetTopProductsFunctionField     = "get_top_products"
	GetTopCustomersFunctionField    = "get_top_customers"
	GetShippingSummaryFunctionField = "get_shipping_summary"
	ParseMetricsShopIDSubFuncField  = "parse_shop_id"
)

type MetricsHandler struct {
	getDashboardUseCase       ports.GetDashboardMetricsUseCase
	getRevenueTrendUseCase    ports.GetRevenueTrendUseCase
	getTopProductsUseCase     ports.GetTopProductsUseCase
	getTopCustomersUseCase    ports.GetTopCustomersUseCase
	getShippingSummaryUseCase ports.GetShippingSummaryUseCase
}

func NewMetricsHandler(
	getDashboardUseCase ports.GetDashboardMetricsUseCase,
	getRevenueTrendUseCase ports.GetRevenueTrendUseCase,
	getTopProductsUseCase ports.GetTopProductsUseCase,
	getTopCustomersUseCase ports.GetTopCustomersUseCase,
	getShippingSummaryUseCase ports.GetShippingSummaryUseCase,
) ports.MetricsHandler {
	return &MetricsHandler{
		getDashboardUseCase:       getDashboardUseCase,
		getRevenueTrendUseCase:    getRevenueTrendUseCase,
		getTopProductsUseCase:     getTopProductsUseCase,
		getTopCustomersUseCase:    getTopCustomersUseCase,
		getShippingSummaryUseCase: getShippingSummaryUseCase,
	}
}

// GetDashboard handles GET /shops/{shop_id}/metrics/dashboard requests.
func (h *MetricsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse optional timezone from query params
	var tz *string
	if tzParam := r.URL.Query().Get("tz"); tzParam != "" {
		tz = &tzParam
	}

	dashboard, err := h.getDashboardUseCase.Execute(ctx, shopID, tz)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetDashboardFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving dashboard metrics")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":     MetricsHandlerField,
		"function": GetDashboardFunctionField,
		"shop_id":  shopID,
	}).Debug("Dashboard metrics retrieved successfully")

	response := responses.NewDashboardResponse(dashboard)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetDashboardFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetRevenueTrend handles GET /shops/{shop_id}/metrics/revenue/trend requests.
func (h *MetricsHandler) GetRevenueTrend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filtersRequest, err := requests.NewMetricsFiltersRequest(r.URL.Query())
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetRevenueTrendFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToMetricsFilters()

	trend, err := h.getRevenueTrendUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetRevenueTrendFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving revenue trend")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":        MetricsHandlerField,
		"function":    GetRevenueTrendFunctionField,
		"shop_id":     shopID,
		"data_points": len(trend),
	}).Debug("Revenue trend retrieved successfully")

	response := responses.NewRevenueTrendResponse(trend)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetRevenueTrendFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetTopProducts handles GET /shops/{shop_id}/metrics/products/top requests.
func (h *MetricsHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filtersRequest, err := requests.NewMetricsFiltersRequest(r.URL.Query())
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopProductsFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToMetricsFilters()

	products, err := h.getTopProductsUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopProductsFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving top products")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":          MetricsHandlerField,
		"function":      GetTopProductsFunctionField,
		"shop_id":       shopID,
		"product_count": len(products),
	}).Debug("Top products retrieved successfully")

	response := responses.NewTopProductsResponse(products)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopProductsFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetTopCustomers handles GET /shops/{shop_id}/metrics/customers/top requests.
func (h *MetricsHandler) GetTopCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filtersRequest, err := requests.NewMetricsFiltersRequest(r.URL.Query())
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopCustomersFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToMetricsFilters()

	customers, err := h.getTopCustomersUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopCustomersFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving top customers")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":           MetricsHandlerField,
		"function":       GetTopCustomersFunctionField,
		"shop_id":        shopID,
		"customer_count": len(customers),
	}).Debug("Top customers retrieved successfully")

	response := responses.NewTopCustomersResponse(customers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetTopCustomersFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// GetShippingSummary handles GET /shops/{shop_id}/metrics/shipping/summary requests.
func (h *MetricsHandler) GetShippingSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filtersRequest, err := requests.NewMetricsFiltersRequest(r.URL.Query())
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetShippingSummaryFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Invalid filter parameters")
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToMetricsFilters()

	summary, err := h.getShippingSummaryUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetShippingSummaryFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving shipping summary")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":     MetricsHandlerField,
		"function": GetShippingSummaryFunctionField,
		"shop_id":  shopID,
	}).Debug("Shipping summary retrieved successfully")

	response := responses.NewShippingSummaryResponse(summary)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": GetShippingSummaryFunctionField,
			"sub_func": "json.Encode",
			"error":    err.Error(),
		}).Error("Error encoding response")
	}
}

// parseShopID extracts and validates shop_id from URL path.
func (h *MetricsHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logs.WithFields(map[string]interface{}{
			"file":     MetricsHandlerField,
			"function": ParseMetricsShopIDSubFuncField,
			"shop_id":  shopIDStr,
			"error":    err,
		}).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}
