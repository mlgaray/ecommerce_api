package http

import (
	"encoding/json"
	"net/http"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	RoleHandlerField            = "role_handler"
	GetAssignableRolesFuncField = "get_assignable"
)

type RoleHandler struct {
	getAssignableRoles ports.GetAssignableRolesUseCase
}

func NewRoleHandler(getAssignableRoles ports.GetAssignableRolesUseCase) *RoleHandler {
	return &RoleHandler{
		getAssignableRoles: getAssignableRoles,
	}
}

func (h *RoleHandler) GetAssignable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roles, err := h.getAssignableRoles.Execute(ctx)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     RoleHandlerField,
			"function": GetAssignableRolesFuncField,
			"error":    err.Error(),
		}).Error("Error retrieving assignable roles")
		httpErrors.HandleError(w, err)
		return
	}

	response := responses.NewListRolesResponse(roles)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
