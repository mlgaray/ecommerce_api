package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	bearerPrefix        = "Bearer "
	authorizationHeader = "Authorization"
)

// Auth log field constants
const (
	AuthMiddlewareField = "auth_middleware"
)

// AuthMiddleware handles JWT token validation and injects claims into context.
type AuthMiddleware struct {
	tokenService ports.TokenService
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
func NewAuthMiddleware(tokenService ports.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Authenticate is a middleware that validates JWT tokens and injects user data into context.
// Use this for routes that require authentication.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get(authorizationHeader)

		// Check if Authorization header is present
		if tokenString == "" {
			logs.WithFields(map[string]interface{}{
				"file":   AuthMiddlewareField,
				"path":   r.URL.Path,
				"method": r.Method,
			}).Warn("Missing authorization header")
			errors.HandleError(w, &errors.UnauthorizedError{Message: "authorization_header_required"})
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(tokenString, bearerPrefix) {
			logs.WithFields(map[string]interface{}{
				"file":   AuthMiddlewareField,
				"path":   r.URL.Path,
				"method": r.Method,
			}).Warn("Invalid token format - missing Bearer prefix")
			errors.HandleError(w, &errors.UnauthorizedError{Message: "invalid_token_format"})
			return
		}

		// Extract token (remove "Bearer " prefix)
		tokenString = strings.TrimPrefix(tokenString, bearerPrefix)
		tokenString = strings.TrimSpace(tokenString)

		// Validate and parse token
		parsedClaims, err := m.tokenService.ValidateAndParseClaims(tokenString)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":   AuthMiddlewareField,
				"path":   r.URL.Path,
				"method": r.Method,
				"error":  err.Error(),
			}).Warn("Token validation failed")
			errors.HandleError(w, err)
			return
		}

		// Inject claims into context
		ctx := r.Context()
		ctx = context.WithValue(ctx, claims.KeyUserID, parsedClaims.UserID)
		ctx = context.WithValue(ctx, claims.KeyEmail, parsedClaims.Email)
		ctx = context.WithValue(ctx, claims.KeyShopIDs, parsedClaims.ShopIDs)

		logs.WithFields(map[string]interface{}{
			"file":     AuthMiddlewareField,
			"path":     r.URL.Path,
			"method":   r.Method,
			"user_id":  parsedClaims.UserID,
			"shop_ids": parsedClaims.ShopIDs,
		}).Debug("User authenticated successfully")

		// Continue with the authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
