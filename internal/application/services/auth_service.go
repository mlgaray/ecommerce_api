package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// func (s *AuthServiceImpl) HashPassword(ctx context.Context, password string) (string, error) {
//	if password == "" {
//		return "", errors.New("password cannot be empty")
//	}
//
//	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//	if err != nil {
//		return "", errors.New("failed to hash password")
//	}
//
//	return string(hashedPassword), nil
//}

func (s *AuthService) ComparePassword(ctx context.Context, hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return &errors.ValidationError{Message: errors.PasswordsCannotBeEmpty}
	}

	if hashedPassword == password {
		return nil // Passwords match
	}

	return &errors.AuthenticationError{Message: errors.InvalidUserCredentials}
}
