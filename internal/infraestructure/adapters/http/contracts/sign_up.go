package contracts

import (
	"regexp"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

type SignUpRequest struct {
	User models.User `json:"user"`
	Shop models.Shop `json:"shop"`
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
