package staff

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type UpdateStaffUseCase struct {
	staffService ports.StaffService
}

func NewUpdateStaffUseCase(staffService ports.StaffService) ports.UpdateStaffUseCase {
	return &UpdateStaffUseCase{staffService: staffService}
}

func (uc *UpdateStaffUseCase) Execute(ctx context.Context, staffID, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error) {
	return uc.staffService.Update(ctx, staffID, shopID, user, roleID, imageBuffer)
}
