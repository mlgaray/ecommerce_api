package role

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetAssignableRolesUseCase struct {
	roleService ports.RoleService
}

func NewGetAssignableRolesUseCase(roleService ports.RoleService) *GetAssignableRolesUseCase {
	return &GetAssignableRolesUseCase{roleService: roleService}
}

func (uc *GetAssignableRolesUseCase) Execute(ctx context.Context) ([]*models.Role, error) {
	return uc.roleService.GetAllAssignable(ctx)
}
