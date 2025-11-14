package contracts

import (
	"regexp"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// SignInRequest represents the sign in request payload
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// emailRegex is a regex pattern for email validation (HTTP layer validation)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._%+-]*[a-zA-Z0-9])?@[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?\.[a-zA-Z]{2,}$`)

// Validate validates HTTP input (format, required fields)
func (r *SignInRequest) Validate() error {
	email := strings.TrimSpace(r.Email)

	// HTTP validation: email required
	if email == "" {
		return &httpErrors.BadRequestError{Message: "email_is_required"}
	}

	// HTTP validation: email format
	if !emailRegex.MatchString(email) {
		return &httpErrors.BadRequestError{Message: "invalid_email_format"}
	}

	// HTTP validation: password required
	if strings.TrimSpace(r.Password) == "" {
		return &httpErrors.BadRequestError{Message: "password_is_required"}
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
