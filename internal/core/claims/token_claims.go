package claims

import "context"

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

// Context keys for authenticated user data.
const (
	KeyUserID  ContextKey = "user_id"
	KeyEmail   ContextKey = "email"
	KeyShopIDs ContextKey = "shop_ids"
)

// TokenClaims represents the parsed JWT token claims.
// Parsing from JWT-specific formats is handled by the auth adapter.
type TokenClaims struct {
	UserID  int
	Email   string
	ShopIDs []int
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
