package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type UserService struct {
	userRepo    ports.UserRepository
	authService ports.AuthService
}

func NewUserService(userRepo ports.UserRepository, authService ports.AuthService) ports.UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *UserService) ValidateCredentials(ctx context.Context, user *models.User, password string) (*models.User, error) {
	err := s.authService.ComparePassword(ctx, user.Password, password)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "compare_password",
			"error":     err.Error(),
		}).Error("error comparing passwords")
		return nil, &errors.UnauthorizedError{Message: errors.InvalidUserCredentials}
	}

	return user, nil
}

func (s *UserService) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return s.userRepo.Create(ctx, user)
}
