package auth

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignInUseCase struct {
	userService       ports.UserService
	tokenService      ports.TokenService
	staffService      ports.StaffService
	permissionService ports.PermissionService
}

func NewSignInUseCase(
	userService ports.UserService,
	tokenService ports.TokenService,
	staffService ports.StaffService,
	permissionService ports.PermissionService,
) ports.SignInUseCase {
	return &SignInUseCase{
		userService:       userService,
		tokenService:      tokenService,
		staffService:      staffService,
		permissionService: permissionService,
	}
}

func (uc *SignInUseCase) Execute(ctx context.Context, user *models.User) (string, error) {
	storedUser, err := uc.userService.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	_, err = uc.userService.ValidateCredentials(ctx, user, storedUser.Password)
	if err != nil {
		return "", err
	}

	// Get shop roles via staff table
	shopRoles, err := uc.staffService.GetShopRolesByUserID(ctx, storedUser.ID)
	if err != nil {
		return "", err
	}

	// Resolve permissions for each role
	for i := range shopRoles {
		shopRoles[i].Permissions = uc.permissionService.GetPermissions(shopRoles[i].Role)
	}

	token, err := uc.tokenService.Generate(ctx, storedUser, shopRoles)
	if err != nil {
		return "", err
	}

	return token, nil
}
