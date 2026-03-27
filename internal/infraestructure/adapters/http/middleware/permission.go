package middleware

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	coreErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

const (
	PermissionMiddlewareField = "permission_middleware"
)

// PermissionMiddleware checks if the authenticated user has the required permission
// for the shop in the request. Uses an in-memory role→permissions map.
type PermissionMiddleware struct {
	permissionService ports.PermissionService
}

// NewPermissionMiddleware creates a new PermissionMiddleware instance.
func NewPermissionMiddleware(permissionService ports.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequirePermission returns a middleware that checks the user's role for the shop
// identified by {shop_id} in the URL path has the given permission.
// Must be applied AFTER AuthMiddleware and ShopOwnershipMiddleware.
func (m *PermissionMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			shopIDStr, exists := vars["shop_id"]
			if !exists {
				// No shop_id in URL — try context (for routes like /products)
				shopID := claims.GetFirstShopIDFromContext(r.Context())
				if shopID == 0 {
					logs.WithFields(map[string]interface{}{
						"file":       PermissionMiddlewareField,
						"path":       r.URL.Path,
						"method":     r.Method,
						"permission": permission,
					}).Warn("Cannot determine shop_id for permission check")
					errors.HandleError(w, &errors.ForbiddenError{Message: coreErrors.InsufficientPermissions})
					return
				}
				m.checkPermission(w, r, next, shopID, permission)
				return
			}

			shopID, err := strconv.Atoi(shopIDStr)
			if err != nil || shopID <= 0 {
				errors.HandleError(w, &errors.BadRequestError{Message: "invalid_shop_id_format"})
				return
			}

			m.checkPermission(w, r, next, shopID, permission)
		})
	}
}

func (m *PermissionMiddleware) checkPermission(w http.ResponseWriter, r *http.Request, next http.Handler, shopID int, permission string) {
	role := claims.GetRoleForShop(r.Context(), shopID)
	if role == "" || !m.permissionService.HasPermission(role, permission) {
		logs.WithFields(map[string]interface{}{
			"file":       PermissionMiddlewareField,
			"path":       r.URL.Path,
			"method":     r.Method,
			"shop_id":    shopID,
			"role":       role,
			"permission": permission,
		}).Warn("Insufficient permissions")
		errors.HandleError(w, &errors.ForbiddenError{Message: coreErrors.InsufficientPermissions})
		return
	}

	next.ServeHTTP(w, r)
}
