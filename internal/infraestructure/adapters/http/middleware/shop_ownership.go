package middleware

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

// Shop ownership log field constants
const (
	ShopOwnershipMiddlewareField = "shop_ownership_middleware"
)

// ShopOwnershipMiddleware verifies that the authenticated user owns the shop
// referenced by {shop_id} in the URL path. Must be applied after AuthMiddleware.
type ShopOwnershipMiddleware struct{}

// NewShopOwnershipMiddleware creates a new ShopOwnershipMiddleware instance.
func NewShopOwnershipMiddleware() *ShopOwnershipMiddleware {
	return &ShopOwnershipMiddleware{}
}

// Authorize is a middleware that checks if the authenticated user owns the shop
// identified by the {shop_id} URL path variable.
func (m *ShopOwnershipMiddleware) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		shopIDStr, exists := vars["shop_id"]
		if !exists {
			// No shop_id in URL — not applicable, pass through
			next.ServeHTTP(w, r)
			return
		}

		shopID, err := strconv.Atoi(shopIDStr)
		if err != nil || shopID <= 0 {
			logs.WithFields(map[string]interface{}{
				"file":    ShopOwnershipMiddlewareField,
				"path":    r.URL.Path,
				"method":  r.Method,
				"shop_id": shopIDStr,
			}).Warn("Invalid shop_id format")
			errors.HandleError(w, &errors.BadRequestError{Message: "invalid_shop_id_format"})
			return
		}

		shopIDs := claims.GetShopIDsFromContext(r.Context())
		if !userOwnsShop(shopIDs, shopID) {
			logs.WithFields(map[string]interface{}{
				"file":       ShopOwnershipMiddlewareField,
				"path":       r.URL.Path,
				"method":     r.Method,
				"shop_id":    shopID,
				"user_shops": shopIDs,
			}).Warn("Access denied to shop")
			errors.HandleError(w, &errors.ForbiddenError{Message: "access_denied_to_shop"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// userOwnsShop checks if the authenticated user owns the given shop.
func userOwnsShop(userShopIDs []int, shopID int) bool {
	for _, id := range userShopIDs {
		if id == shopID {
			return true
		}
	}
	return false
}
