package requests

import (
	"regexp"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// slugRegex validates slug format: lowercase alphanumeric + hyphens,
// no leading/trailing/consecutive hyphens.
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const (
	slugMinLength = 3
	slugMaxLength = 63
)

// ValidateSlug validates that a slug meets format requirements:
// - 3 to 63 characters
// - Lowercase alphanumeric and hyphens only
// - No leading, trailing, or consecutive hyphens
func ValidateSlug(slug string) error {
	if len(slug) < slugMinLength || len(slug) > slugMaxLength || !slugRegex.MatchString(slug) {
		return &httpErrors.BadRequestError{Message: domainErrors.InvalidSlugFormat}
	}
	return nil
}
