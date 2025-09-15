package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type TokenService interface {
	Generate(ctx context.Context, user *models.User) (string, error)
	// ValidateToken(ctx context.Context, token string) (*entities.User, error)
	// RefreshToken(ctx context.Context, token string) (string, error)
	// GetTokenExpiration() time.Duration
}
