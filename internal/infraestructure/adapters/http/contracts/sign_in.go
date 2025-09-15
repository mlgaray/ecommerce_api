package contracts

import (
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// SignInRequest represents the sign in request payload
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// emailRegex is a regex pattern for email validation

// Validate validates the sign in request
func (r *SignInRequest) Validate() error {
	email := strings.TrimSpace(r.Email)
	if email == "" {
		return &errors.BadRequestError{Message: "email_is_required"}
	}
	if !emailRegex.MatchString(email) {
		return &errors.BadRequestError{Message: "invalid_email_format"}
	}
	if strings.TrimSpace(r.Password) == "" {
		return &errors.BadRequestError{Message: "password_is_required"}
	}
	return nil
}

// ToUser converts the request to a User model
func (r *SignInRequest) ToUser() *models.User {
	return &models.User{
		Email:    strings.TrimSpace(r.Email),
		Password: strings.TrimSpace(r.Password),
	}
}

// SignInResponse represents the successful sign in response
type SignInResponse struct {
	Token string `json:"token"`
}
