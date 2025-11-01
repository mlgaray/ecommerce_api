package contracts

import (
	"regexp"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type SignUpRequest struct {
	User models.User `json:"user"`
	Shop models.Shop `json:"shop"`
}

// emailRegex is a regex pattern for email validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._%+-]*[a-zA-Z0-9])?@[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?\.[a-zA-Z]{2,}$`)

func (r *SignUpRequest) Validate() error {
	if err := r.validateUser(); err != nil {
		return err
	}
	return r.validateShop()
}

func (r *SignUpRequest) validateUser() error {
	if strings.TrimSpace(r.User.Name) == "" {
		return &errors.BadRequestError{Message: "user_name_is_required"}
	}
	if strings.TrimSpace(r.User.LastName) == "" {
		return &errors.BadRequestError{Message: "user_last_name_is_required"}
	}
	if strings.TrimSpace(r.User.Email) == "" {
		return &errors.BadRequestError{Message: "user_email_is_required"}
	}
	if !emailRegex.MatchString(strings.TrimSpace(r.User.Email)) {
		return &errors.BadRequestError{Message: "invalid_email_format"}
	}
	if strings.TrimSpace(r.User.Phone) == "" {
		return &errors.BadRequestError{Message: "user_phone_is_required"}
	}
	if strings.TrimSpace(r.User.Password) == "" {
		return &errors.BadRequestError{Message: "user_password_is_required"}
	}
	return nil
}

func (r *SignUpRequest) validateShop() error {
	if strings.TrimSpace(r.Shop.Name) == "" {
		return &errors.BadRequestError{Message: "shop_name_is_required"}
	}
	if strings.TrimSpace(r.Shop.Slug) == "" {
		return &errors.BadRequestError{Message: "shop_slug_is_required"}
	}
	if strings.TrimSpace(r.Shop.Email) == "" {
		return &errors.BadRequestError{Message: "shop_email_is_required"}
	}
	if strings.TrimSpace(r.Shop.Phone) == "" {
		return &errors.BadRequestError{Message: "shop_phone_is_required"}
	}
	return nil
}
