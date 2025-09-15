package auth

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignUpUseCase struct {
	signUpService ports.SignUpService
}

func (uc *SignUpUseCase) Execute(ctx context.Context, user *models.User, shop *models.Shop) error {
	_, err := uc.signUpService.SignUp(ctx, user, shop)
	return err
}

func NewSignUpUseCase(signUpService ports.SignUpService) ports.SignUpUseCase {
	return &SignUpUseCase{
		signUpService: signUpService,
	}
}
