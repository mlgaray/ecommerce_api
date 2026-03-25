package staff

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetStaffByIDUseCase struct {
	staffService ports.StaffService
}

func NewGetStaffByIDUseCase(staffService ports.StaffService) ports.GetStaffByIDUseCase {
	return &GetStaffByIDUseCase{staffService: staffService}
}

func (uc *GetStaffByIDUseCase) Execute(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	return uc.staffService.GetByID(ctx, staffID, shopID)
}
