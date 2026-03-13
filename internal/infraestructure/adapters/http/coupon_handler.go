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

// Coupon handler log field constants
const (
	CouponHandlerField         = "coupon_handler"
	CreateCouponFunctionField  = "create"
	UpdateCouponFunctionField  = "update"
	DeleteCouponFunctionField  = "delete"
	GetCouponByIDFunctionField = "get_by_id"
	GetAllCouponsFunctionField = "get_all"
	ParseCouponShopIDSubField  = "parse_shop_id"
	ParseCouponIDSubField      = "parse_coupon_id"
)

type CouponHandler struct {
	createCouponUseCase  ports.CreateCouponUseCase
	updateCouponUseCase  ports.UpdateCouponUseCase
	deleteCouponUseCase  ports.DeleteCouponUseCase
	getCouponByIDUseCase ports.GetCouponByIDUseCase
	getAllCouponsUseCase ports.GetAllCouponsByShopIDUseCase
}

func NewCouponHandler(
	createCouponUseCase ports.CreateCouponUseCase,
	updateCouponUseCase ports.UpdateCouponUseCase,
	deleteCouponUseCase ports.DeleteCouponUseCase,
	getCouponByIDUseCase ports.GetCouponByIDUseCase,
	getAllCouponsUseCase ports.GetAllCouponsByShopIDUseCase,
) ports.CouponHandler {
	return &CouponHandler{
		createCouponUseCase:  createCouponUseCase,
		updateCouponUseCase:  updateCouponUseCase,
		deleteCouponUseCase:  deleteCouponUseCase,
		getCouponByIDUseCase: getCouponByIDUseCase,
		getAllCouponsUseCase: getAllCouponsUseCase,
	}
}

// Create handles POST /shops/{shop_id}/coupons requests.
func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	var request requests.CreateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CouponHandlerField,
			"function": CreateCouponFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Warn("Invalid JSON body")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_json_body"})
		return
	}

	if err := request.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	coupon := request.ToModel()

	created, err := h.createCouponUseCase.Execute(ctx, coupon, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CouponHandlerField,
			"function": CreateCouponFunctionField,
			"shop_id":  shopID,
			"code":     coupon.Code,
			"error":    err.Error(),
		}).Error("Error creating coupon")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":      CouponHandlerField,
		"function":  CreateCouponFunctionField,
		"shop_id":   shopID,
		"coupon_id": created.ID,
		"code":      created.Code,
	}).Info("Coupon created successfully")

	response := responses.CreateCouponResponse{
		Coupon: responses.CouponResponseFromModel(created),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// Update handles PUT /shops/{shop_id}/coupons/{coupon_id} requests.
func (h *CouponHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	couponID, err := h.parseCouponID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	var request requests.UpdateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponHandlerField,
			"function":  UpdateCouponFunctionField,
			"shop_id":   shopID,
			"coupon_id": couponID,
			"error":     err.Error(),
		}).Warn("Invalid JSON body")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_json_body"})
		return
	}

	if err := request.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	coupon := request.ToModel()

	updated, err := h.updateCouponUseCase.Execute(ctx, couponID, shopID, coupon)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponHandlerField,
			"function":  UpdateCouponFunctionField,
			"shop_id":   shopID,
			"coupon_id": couponID,
			"error":     err.Error(),
		}).Error("Error updating coupon")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":      CouponHandlerField,
		"function":  UpdateCouponFunctionField,
		"shop_id":   shopID,
		"coupon_id": couponID,
	}).Info("Coupon updated successfully")

	response := responses.UpdateCouponResponse{
		Coupon: responses.CouponResponseFromModel(updated),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Delete handles DELETE /shops/{shop_id}/coupons/{coupon_id} requests.
func (h *CouponHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	couponID, err := h.parseCouponID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := h.deleteCouponUseCase.Execute(ctx, couponID, shopID); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponHandlerField,
			"function":  DeleteCouponFunctionField,
			"shop_id":   shopID,
			"coupon_id": couponID,
			"error":     err.Error(),
		}).Error("Error deleting coupon")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":      CouponHandlerField,
		"function":  DeleteCouponFunctionField,
		"shop_id":   shopID,
		"coupon_id": couponID,
	}).Info("Coupon deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

// GetByID handles GET /shops/{shop_id}/coupons/{coupon_id} requests.
func (h *CouponHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	couponID, err := h.parseCouponID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	coupon, err := h.getCouponByIDUseCase.Execute(ctx, couponID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      CouponHandlerField,
			"function":  GetCouponByIDFunctionField,
			"shop_id":   shopID,
			"coupon_id": couponID,
			"error":     err.Error(),
		}).Error("Error retrieving coupon")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.GetCouponResponse{
		Coupon: responses.CouponResponseFromModel(coupon),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetAll handles GET /shops/{shop_id}/coupons requests.
func (h *CouponHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	queryParams := r.URL.Query()
	filtersRequest, err := requests.NewCouponFiltersRequest(queryParams)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToCouponFilters()

	coupons, nextCursor, hasMore, totalCount, err := h.getAllCouponsUseCase.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     CouponHandlerField,
			"function": GetAllCouponsFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving coupons")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.NewListCouponsResponse(coupons, nextCursor, hasMore, totalCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// parseShopID extracts and validates shop_id from URL path.
func (h *CouponHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logFields := map[string]interface{}{
			"file":     CouponHandlerField,
			"function": ParseCouponShopIDSubField,
			"shop_id":  shopIDStr,
		}
		if err != nil {
			logFields["error"] = err.Error()
		}
		logs.WithFields(logFields).Error("Invalid shop_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}

	return shopID, nil
}

// parseCouponID extracts and validates coupon_id from URL path.
func (h *CouponHandler) parseCouponID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	couponIDStr := vars["coupon_id"]

	couponID, err := strconv.Atoi(couponIDStr)
	if err != nil || couponID <= 0 {
		logFields := map[string]interface{}{
			"file":      CouponHandlerField,
			"function":  ParseCouponIDSubField,
			"coupon_id": couponIDStr,
		}
		if err != nil {
			logFields["error"] = err.Error()
		}
		logs.WithFields(logFields).Error("Invalid coupon_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_coupon_id_format"}
	}

	return couponID, nil
}
