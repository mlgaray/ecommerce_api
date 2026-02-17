package requests

import (
	"regexp"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Sign In
// =============================================================================

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

// =============================================================================
// Sign Up
// =============================================================================

// SignUpRequest represents the HTTP request for user registration.
type SignUpRequest struct {
	User UserRequest       `json:"user"`
	Shop SignUpShopRequest `json:"shop"`
}

// UserRequest represents user data in sign up requests.
type UserRequest struct {
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// SignUpShopRequest represents shop data in sign up requests.
type SignUpShopRequest struct {
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// signUpEmailRegex is a regex pattern for email validation (HTTP layer validation)
var signUpEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._%+-]*[a-zA-Z0-9])?@[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?\.[a-zA-Z]{2,}$`)

func (r *SignUpRequest) Validate() error {
	if err := r.validateUser(); err != nil {
		return err
	}
	return r.validateShop()
}

// validateUser validates HTTP input for user fields
func (r *SignUpRequest) validateUser() error {
	// HTTP validation: user name required
	if strings.TrimSpace(r.User.Name) == "" {
		return &httpErrors.BadRequestError{Message: "user_name_is_required"}
	}

	// HTTP validation: user last name required
	if strings.TrimSpace(r.User.LastName) == "" {
		return &httpErrors.BadRequestError{Message: "user_last_name_is_required"}
	}

	// HTTP validation: user email required
	if strings.TrimSpace(r.User.Email) == "" {
		return &httpErrors.BadRequestError{Message: "user_email_is_required"}
	}

	// HTTP validation: email format
	if !signUpEmailRegex.MatchString(strings.TrimSpace(r.User.Email)) {
		return &httpErrors.BadRequestError{Message: "invalid_email_format"}
	}

	// HTTP validation: user phone required
	if strings.TrimSpace(r.User.Phone) == "" {
		return &httpErrors.BadRequestError{Message: "user_phone_is_required"}
	}

	// HTTP validation: user password required
	if strings.TrimSpace(r.User.Password) == "" {
		return &httpErrors.BadRequestError{Message: "user_password_is_required"}
	}

	return nil
}

// validateShop validates HTTP input for shop fields
func (r *SignUpRequest) validateShop() error {
	// HTTP validation: shop name required
	if strings.TrimSpace(r.Shop.Name) == "" {
		return &httpErrors.BadRequestError{Message: "shop_name_is_required"}
	}

	// HTTP validation: shop slug required
	if strings.TrimSpace(r.Shop.Slug) == "" {
		return &httpErrors.BadRequestError{Message: "shop_slug_is_required"}
	}

	// HTTP validation: shop email required
	if strings.TrimSpace(r.Shop.Email) == "" {
		return &httpErrors.BadRequestError{Message: "shop_email_is_required"}
	}

	// HTTP validation: shop phone required
	if strings.TrimSpace(r.Shop.Phone) == "" {
		return &httpErrors.BadRequestError{Message: "shop_phone_is_required"}
	}

	return nil
}

// ToUserModel converts the sign up user to a domain model.
func (r *SignUpRequest) ToUserModel() *models.User {
	return &models.User{
		Name:     r.User.Name,
		LastName: r.User.LastName,
		Email:    r.User.Email,
		Password: r.User.Password,
		Phone:    r.User.Phone,
	}
}

// ToShopModel converts the sign up shop to a domain model.
func (r *SignUpRequest) ToShopModel() *models.Shop {
	return &models.Shop{
		Name:  r.Shop.Name,
		Slug:  r.Shop.Slug,
		Email: r.Shop.Email,
		Phone: r.Shop.Phone,
	}
}
