package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type SignupService struct {
	signupRepo  ports.SignupRepository
	authService ports.AuthService
}

func NewSignupService(signupRepo ports.SignupRepository, authService ports.AuthService) ports.SignUpService {
	return &SignupService{
		signupRepo:  signupRepo,
		authService: authService,
	}
}

func (s *SignupService) SignUp(ctx context.Context, user *models.User, shop *models.Shop) (*models.User, error) {
	hashedPassword, err := s.authService.HashPassword(ctx, user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = hashedPassword

	return s.signupRepo.CreateUserWithShop(ctx, user, shop)
}
