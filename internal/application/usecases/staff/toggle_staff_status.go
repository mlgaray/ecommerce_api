package staff

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type ToggleStaffStatusUseCase struct {
	staffService ports.StaffService
}

func NewToggleStaffStatusUseCase(staffService ports.StaffService) ports.ToggleStaffStatusUseCase {
	return &ToggleStaffStatusUseCase{staffService: staffService}
}

func (uc *ToggleStaffStatusUseCase) Execute(ctx context.Context, staffID, shopID int) (*models.Staff, error) {
	return uc.staffService.ToggleStatus(ctx, staffID, shopID)
}
