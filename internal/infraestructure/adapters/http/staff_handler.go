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

const (
	StaffHandlerField              = "staff_handler"
	CreateStaffFunctionField       = "create"
	GetAllStaffFunctionField       = "get_all"
	GetStaffByIDFunctionField      = "get_by_id"
	UpdateStaffFunctionField       = "update"
	DeleteStaffFunctionField       = "delete"
	ToggleStaffStatusFunctionField = "toggle_status"
	ParseStaffShopIDSubField       = "parse_shop_id"
	ParseStaffIDSubField           = "parse_staff_id"
)

type StaffHandler struct {
	createStaff       ports.CreateStaffUseCase
	getAllStaff       ports.GetAllStaffByShopIDUseCase
	getStaffByID      ports.GetStaffByIDUseCase
	updateStaff       ports.UpdateStaffUseCase
	deleteStaff       ports.DeleteStaffUseCase
	toggleStaffStatus ports.ToggleStaffStatusUseCase
}

func NewStaffHandler(
	createStaff ports.CreateStaffUseCase,
	getAllStaff ports.GetAllStaffByShopIDUseCase,
	getStaffByID ports.GetStaffByIDUseCase,
	updateStaff ports.UpdateStaffUseCase,
	deleteStaff ports.DeleteStaffUseCase,
	toggleStaffStatus ports.ToggleStaffStatusUseCase,
) *StaffHandler {
	return &StaffHandler{
		createStaff:       createStaff,
		getAllStaff:       getAllStaff,
		getStaffByID:      getStaffByID,
		updateStaff:       updateStaff,
		deleteStaff:       deleteStaff,
		toggleStaffStatus: toggleStaffStatus,
	}
}

// Create handles POST /shops/{shop_id}/staff requests.
func (h *StaffHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	err = r.ParseMultipartForm(4 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": CreateStaffFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	request, err := requests.NewCreateStaffRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": CreateStaffFunctionField,
			"sub_func": "NewCreateStaffRequest",
			"error":    err.Error(),
		}).Error("Error building staff request")
		httpErrors.HandleError(w, err)
		return
	}

	if err := request.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	imageBuffer, err := request.ToImageBuffer()
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staff, err := h.createStaff.Execute(ctx, shopID, request.ToUserModel(), request.RoleID, imageBuffer)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": CreateStaffFunctionField,
			"shop_id":  shopID,
			"email":    request.Email,
			"error":    err.Error(),
		}).Error("Error creating staff")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":     StaffHandlerField,
		"function": CreateStaffFunctionField,
		"shop_id":  shopID,
		"staff_id": staff.ID,
	}).Info("Staff created successfully")

	response := responses.CreateStaffResponse{
		Staff: responses.StaffResponseFromModel(staff),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// GetAll handles GET /shops/{shop_id}/staff requests.
func (h *StaffHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	queryParams := r.URL.Query()
	filtersRequest, err := requests.NewStaffFiltersRequest(queryParams)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	if err := filtersRequest.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	filters := filtersRequest.ToStaffFilters()

	staffList, nextCursor, hasMore, totalCount, err := h.getAllStaff.Execute(ctx, shopID, filters)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": GetAllStaffFunctionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error retrieving staff")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.NewListStaffResponse(staffList, nextCursor, hasMore, totalCount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetByID handles GET /shops/{shop_id}/staff/{staff_id} requests.
func (h *StaffHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staffID, err := h.parseStaffID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staff, err := h.getStaffByID.Execute(ctx, staffID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": GetStaffByIDFunctionField,
			"shop_id":  shopID,
			"staff_id": staffID,
			"error":    err.Error(),
		}).Error("Error retrieving staff")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.GetStaffResponse{
		Staff: responses.StaffResponseFromModel(staff),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Update handles PUT /shops/{shop_id}/staff/{staff_id} requests.
func (h *StaffHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staffID, err := h.parseStaffID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	err = r.ParseMultipartForm(4 << 20)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": UpdateStaffFunctionField,
			"sub_func": "r.ParseMultipartForm",
			"error":    err.Error(),
		}).Error("Error parsing multipart form")
		httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
		return
	}

	request, err := requests.NewUpdateStaffRequest(r)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": UpdateStaffFunctionField,
			"sub_func": "NewUpdateStaffRequest",
			"error":    err.Error(),
		}).Error("Error building staff update request")
		httpErrors.HandleError(w, err)
		return
	}

	if err := request.Validate(); err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	imageBuffer, err := request.ToImageBuffer()
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staff, err := h.updateStaff.Execute(ctx, staffID, shopID, request.ToUserModel(), request.RoleID, imageBuffer)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": UpdateStaffFunctionField,
			"shop_id":  shopID,
			"staff_id": staffID,
			"error":    err.Error(),
		}).Error("Error updating staff")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":     StaffHandlerField,
		"function": UpdateStaffFunctionField,
		"shop_id":  shopID,
		"staff_id": staffID,
	}).Info("Staff updated successfully")

	response := responses.UpdateStaffResponse{
		Staff: responses.StaffResponseFromModel(staff),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Delete handles DELETE /shops/{shop_id}/staff/{staff_id} requests.
func (h *StaffHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staffID, err := h.parseStaffID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	err = h.deleteStaff.Execute(ctx, staffID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": DeleteStaffFunctionField,
			"shop_id":  shopID,
			"staff_id": staffID,
			"error":    err.Error(),
		}).Error("Error deleting staff")
		httpErrors.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleStatus handles PATCH /shops/{shop_id}/staff/{staff_id}/status requests.
func (h *StaffHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	shopID, err := h.parseShopID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staffID, err := h.parseStaffID(r)
	if err != nil {
		httpErrors.HandleError(w, err)
		return
	}

	staff, err := h.toggleStaffStatus.Execute(ctx, staffID, shopID)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     StaffHandlerField,
			"function": ToggleStaffStatusFunctionField,
			"shop_id":  shopID,
			"staff_id": staffID,
			"error":    err.Error(),
		}).Error("Error toggling staff status")
		httpErrors.HandleError(w, err)
		return
	}

	logs.WithFields(map[string]interface{}{
		"file":      StaffHandlerField,
		"function":  ToggleStaffStatusFunctionField,
		"shop_id":   shopID,
		"staff_id":  staffID,
		"is_active": staff.IsActive,
	}).Info("Staff status toggled successfully")

	response := responses.ToggleStaffStatusResponse{
		Staff: responses.StaffResponseFromModel(staff),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *StaffHandler) parseShopID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	shopIDStr := vars["shop_id"]

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil || shopID <= 0 {
		logFields := map[string]interface{}{
			"file":     StaffHandlerField,
			"function": ParseStaffShopIDSubField,
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

func (h *StaffHandler) parseStaffID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	staffIDStr := vars["staff_id"]

	staffID, err := strconv.Atoi(staffIDStr)
	if err != nil || staffID <= 0 {
		logFields := map[string]interface{}{
			"file":     StaffHandlerField,
			"function": ParseStaffIDSubField,
			"staff_id": staffIDStr,
		}
		if err != nil {
			logFields["error"] = err.Error()
		}
		logs.WithFields(logFields).Error("Invalid staff_id parameter")
		return 0, &httpErrors.BadRequestError{Message: "invalid_staff_id_format"}
	}

	return staffID, nil
}
