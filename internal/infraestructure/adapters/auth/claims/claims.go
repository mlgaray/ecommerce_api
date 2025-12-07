package claims

import (
	"context"
	"encoding/json"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

// Context keys for authenticated user data.
const (
	KeyUserID  ContextKey = "user_id"
	KeyEmail   ContextKey = "email"
	KeyShopIDs ContextKey = "shop_ids"
)

// TokenClaims represents the parsed JWT token claims.
type TokenClaims struct {
	UserID  int
	Email   string
	ShopIDs []int
}

// NewTokenClaims creates TokenClaims from a map of JWT claims.
// Returns AuthenticationError if claims are invalid or missing required fields.
func NewTokenClaims(claims map[string]interface{}) (*TokenClaims, error) {
	user, err := parseUserFromClaims(claims)
	if err != nil {
		return nil, err
	}

	shopIDs := parseShopIDsFromClaims(claims)

	return &TokenClaims{
		UserID:  user.ID,
		Email:   user.Email,
		ShopIDs: shopIDs,
	}, nil
}

func parseUserFromClaims(claims map[string]interface{}) (*models.User, error) {
	userJSON, ok := claims["user"].(string)
	if !ok {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, &errors.AuthenticationError{Message: errors.TokenInvalid}
	}

	return &user, nil
}

func parseShopIDsFromClaims(claims map[string]interface{}) []int {
	var shopIDs []int
	if rawShopIDs, ok := claims["shop_ids"].([]interface{}); ok {
		for _, id := range rawShopIDs {
			if floatID, ok := id.(float64); ok {
				shopIDs = append(shopIDs, int(floatID))
			}
		}
	}
	return shopIDs
}

// GetUserIDFromContext extracts user_id from context.
// Returns 0 if not found.
func GetUserIDFromContext(ctx context.Context) int {
	if userID, ok := ctx.Value(KeyUserID).(int); ok {
		return userID
	}
	return 0
}

// GetEmailFromContext extracts email from context.
// Returns empty string if not found.
func GetEmailFromContext(ctx context.Context) string {
	if email, ok := ctx.Value(KeyEmail).(string); ok {
		return email
	}
	return ""
}

// GetShopIDsFromContext extracts shop_ids from context.
// Returns nil if not found.
func GetShopIDsFromContext(ctx context.Context) []int {
	if shopIDs, ok := ctx.Value(KeyShopIDs).([]int); ok {
		return shopIDs
	}
	return nil
}

// GetFirstShopIDFromContext extracts the first shop_id from context.
// Returns 0 if not found or empty.
// Useful when user has only one shop (current implementation).
func GetFirstShopIDFromContext(ctx context.Context) int {
	shopIDs := GetShopIDsFromContext(ctx)
	if len(shopIDs) > 0 {
		return shopIDs[0]
	}
	return 0
}
