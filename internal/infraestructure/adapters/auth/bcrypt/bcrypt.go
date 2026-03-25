package bcrypt

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) HashPassword(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", &errors.ValidationError{Message: errors.PasswordsCannotBeEmpty}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", &errors.ValidationError{Message: "failed_to_hash_password"}
	}

	return string(hashedPassword), nil
}

func (s *AuthService) ComparePassword(ctx context.Context, hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return &errors.ValidationError{Message: errors.PasswordsCannotBeEmpty}
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return &errors.AuthenticationError{Message: errors.InvalidUserCredentials}
	}

	return nil
}
