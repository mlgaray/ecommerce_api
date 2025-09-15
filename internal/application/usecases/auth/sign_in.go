package auth

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignInUseCase struct {
	userService  ports.UserService
	tokenService ports.TokenService
}

func NewSignInUseCase(userService ports.UserService, tokenService ports.TokenService) ports.SignInUseCase {
	return &SignInUseCase{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (uc *SignInUseCase) Execute(ctx context.Context, user *models.User) (string, error) {
	_user, err := uc.userService.GetByEmail(ctx, user.Email)
	if err != nil {
		return "", err
	}

	authenticatedUser, err := uc.userService.ValidateCredentials(ctx, user, _user.Password)
	if err != nil {
		return "", err
	}

	token, err := uc.tokenService.Generate(ctx, authenticatedUser)
	if err != nil {
		return "", err
	}

	return token, nil
}
