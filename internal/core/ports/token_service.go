package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	authclaims "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
)

type TokenService interface {
	// Generate creates a new JWT token with user data and shop IDs.
	Generate(ctx context.Context, user *models.User, shopIDs []int) (string, error)
	// ValidateAndParseClaims validates a JWT token and returns the parsed claims.
	// Returns AuthenticationError if token is invalid, expired, or malformed.
	ValidateAndParseClaims(token string) (*authclaims.TokenClaims, error)
}
