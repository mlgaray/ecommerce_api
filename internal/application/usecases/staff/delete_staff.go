package staff

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type DeleteStaffUseCase struct {
	staffService ports.StaffService
}

func NewDeleteStaffUseCase(staffService ports.StaffService) ports.DeleteStaffUseCase {
	return &DeleteStaffUseCase{staffService: staffService}
}

func (uc *DeleteStaffUseCase) Execute(ctx context.Context, staffID, shopID int) error {
	return uc.staffService.Delete(ctx, staffID, shopID)
}
