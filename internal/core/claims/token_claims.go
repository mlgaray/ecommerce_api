package claims

import "context"

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

// Context keys for authenticated user data.
const (
	KeyUserID    ContextKey = "user_id"
	KeyEmail     ContextKey = "email"
	KeyShopIDs   ContextKey = "shop_ids"
	KeyShopRoles ContextKey = "shop_roles"
)

// ShopRole represents a user's role and permissions in a specific shop.
// Embedded in the JWT token so the frontend knows what the user can do.
type ShopRole struct {
	ShopID      int      `json:"id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// TokenClaims represents the parsed JWT token claims.
// Parsing from JWT-specific formats is handled by the auth adapter.
type TokenClaims struct {
	UserID    int
	Email     string
	ShopRoles []ShopRole
}

// ShopIDs derives a flat list of shop IDs from ShopRoles.
// Backward compatible with code that only needs shop IDs (e.g., ShopOwnershipMiddleware).
func (tc *TokenClaims) ShopIDs() []int {
	ids := make([]int, len(tc.ShopRoles))
	for i, sr := range tc.ShopRoles {
		ids[i] = sr.ShopID
	}
	return ids
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
func GetFirstShopIDFromContext(ctx context.Context) int {
	shopIDs := GetShopIDsFromContext(ctx)
	if len(shopIDs) > 0 {
		return shopIDs[0]
	}
	return 0
}

// GetShopRolesFromContext extracts shop_roles from context.
// Returns nil if not found.
func GetShopRolesFromContext(ctx context.Context) []ShopRole {
	if shopRoles, ok := ctx.Value(KeyShopRoles).([]ShopRole); ok {
		return shopRoles
	}
	return nil
}

// GetRoleForShop returns the user's role for a specific shop.
// Returns empty string if the user has no role in that shop.
func GetRoleForShop(ctx context.Context, shopID int) string {
	for _, sr := range GetShopRolesFromContext(ctx) {
		if sr.ShopID == shopID {
			return sr.Role
		}
	}
	return ""
}
