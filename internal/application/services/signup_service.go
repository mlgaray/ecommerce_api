package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignupService struct {
	signupRepo ports.SignupRepository
}

func NewSignupService(signupRepo ports.SignupRepository) ports.SignUpService {
	return &SignupService{
		signupRepo: signupRepo,
	}
}

func (s *SignupService) SignUp(ctx context.Context, user *models.User, shop *models.Shop) (*models.User, error) {
	user.IsActive = true
	return s.signupRepo.CreateUserWithShop(ctx, user, shop)
}
