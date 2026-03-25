package ports

import (
	"context"

	authclaims "github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type TokenService interface {
	// Generate creates a new JWT token with user data and shop roles.
	Generate(ctx context.Context, user *models.User, shopRoles []authclaims.ShopRole) (string, error)
	// ValidateAndParseClaims validates a JWT token and returns the parsed claims.
	// Returns AuthenticationError if token is invalid, expired, or malformed.
	ValidateAndParseClaims(token string) (*authclaims.TokenClaims, error)
}
