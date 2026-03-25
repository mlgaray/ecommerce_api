package staff

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type CreateStaffUseCase struct {
	staffService ports.StaffService
}

func NewCreateStaffUseCase(staffService ports.StaffService) ports.CreateStaffUseCase {
	return &CreateStaffUseCase{staffService: staffService}
}

func (uc *CreateStaffUseCase) Execute(ctx context.Context, shopID int, user *models.User, roleID int, imageBuffer []byte) (*models.Staff, error) {
	return uc.staffService.Create(ctx, shopID, user, roleID, imageBuffer)
}
