package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type RoleServiceImpl struct {
	roleRepository ports.RoleRepository
}

func NewRoleService(roleRepository ports.RoleRepository) *RoleServiceImpl {
	return &RoleServiceImpl{roleRepository: roleRepository}
}

func (s *RoleServiceImpl) GetAllAssignable(ctx context.Context) ([]*models.Role, error) {
	return s.roleRepository.GetAllAssignable(ctx)
}
