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

// Shop handler log field constants
const (
	ShopHandlerField         = "shop_handler"
	GetShopByIDFunctionField = "get_by_id"
	UpdateShopFunctionField  = "update"
)

type ShopHandler struct {
	getShopByID ports.GetShopByIDUseCase
	updateShop  ports.UpdateShopUseCase
}

func NewShopHandler(
	getShopByID ports.GetShopByIDUseCase,
	updateShop ports.UpdateShopUseCase,
) ports.ShopHandler {
	return &ShopHandler{
		getShopByID: getShopByID,
		updateShop:  updateShop,
	}
}

func (h *ShopHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse shop_id from URL
	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	shop, err := h.getShopByID.Execute(ctx, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": GetShopByIDFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving shop")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.GetShopResponse{
		Shop: responses.ShopResponseFromModel(shop),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *ShopHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		return 0, &httpErrors.BadRequestError{Message: "invalid_shop_id_format"}
	}
	return shopID, nil
}

//nolint:gocyclo // Complexity is intentional for readability - sequential request handling
func (h *ShopHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse shop_id from URL
	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	// Parse multipart form (7MB limit - allows logo + cover of 3MB each + shop data)
	err = r.ParseMultipartForm(7 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": UpdateShopFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	// Build shop update request from multipart form
	request, err := requests.NewShopUpdateRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": UpdateShopFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error building shop update request")
		httpErrors.HandleError(w, err)
		return
	}

	// DEBUG: Log payment methods to understand incoming data
	for i, pm := range request.Shop.PaymentMethods {
		logs.WithFields(map[string]interface{}{
			"file":               ShopHandlerField,
			"function":           UpdateShopFunctionField,
			"shop_id":            shopID,
			"payment_method_idx": i,
			"pm_id":              pm.ID,
			"pm_is_active":       pm.IsActive,
			"has_transfer":       pm.TransferConfig != nil,
			"has_mercadopago":    pm.MercadoPagoConfig != nil,
		}).Debug("Payment method received")

		if pm.MercadoPagoConfig != nil {
			logs.WithFields(map[string]interface{}{
				"file":            ShopHandlerField,
				"function":        UpdateShopFunctionField,
				"shop_id":         shopID,
				"mp_has_token":    pm.MercadoPagoConfig.AccessToken != "",
				"mp_has_pubkey":   pm.MercadoPagoConfig.PublicKey != "",
				"mp_access_token": pm.MercadoPagoConfig.AccessToken,
				"mp_public_key":   pm.MercadoPagoConfig.PublicKey,
			}).Debug("MercadoPago config details")
		}
	}

	// Validate request
	if err := request.Validate(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":      ShopHandlerField,
			"function":  UpdateShopFunctionField,
			"shop_id":   shopID,
			"shop_name": request.Shop.Name,
			"error":     err.Error(),
		}).Warn("Shop update validation failed")
		httpErrors.HandleError(w, err)
		return
	}

	// Convert new images to buffers
	logoBuffer, err := request.ToLogoBuffer()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": UpdateShopFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error converting logo to buffer")
		httpErrors.HandleError(w, err)
		return
	}

	coverBuffer, err := request.ToCoverBuffer()
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": UpdateShopFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error converting cover to buffer")
		httpErrors.HandleError(w, err)
		return
	}

	// Execute use case
	err = h.updateShop.Execute(ctx, shopID, request.ToModel(), logoBuffer, coverBuffer)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ShopHandlerField,
			"function": UpdateShopFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error updating shop")
		httpErrors.HandleError(w, err)
		return
	}

	// Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
