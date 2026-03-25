package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func contextWithShopRoles(shopRoles []claims.ShopRole) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, claims.KeyShopRoles, shopRoles)
	// Also set shop IDs for the fallback path (GetFirstShopIDFromContext)
	ids := make([]int, len(shopRoles))
	for i, sr := range shopRoles {
		ids[i] = sr.ShopID
	}
	ctx = context.WithValue(ctx, claims.KeyShopIDs, ids)
	return ctx
}

func TestNewPermissionMiddleware(t *testing.T) {
	t.Run("creates new permission middleware instance", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)

		// Act
		mw := NewPermissionMiddleware(mockPermissionService)

		// Assert
		assert.NotNil(t, mw)
		assert.Equal(t, mockPermissionService, mw.permissionService)
	})
}

func TestPermissionMiddleware_RequirePermission(t *testing.T) {
	t.Run("when user has permission for shop then passes through with 200", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		mockPermissionService.EXPECT().
			HasPermission("owner", "manage_staff").
			Return(true)

		shopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/1/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		mw.RequirePermission("manage_staff")(okHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when user lacks permission for shop then returns 403", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		mockPermissionService.EXPECT().
			HasPermission("staff", "manage_staff").
			Return(false)

		shopRoles := []claims.ShopRole{{ShopID: 1, Role: "staff"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/1/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_staff")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "insufficient_permissions", response["error"])
	})

	t.Run("when user has no role for that shop then returns 403", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		// User has role for shop 99, but request is for shop 1
		shopRoles := []claims.ShopRole{{ShopID: 99, Role: "owner"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/1/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_staff")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "insufficient_permissions", response["error"])
	})

	t.Run("when shop_id is invalid format then returns 400", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		shopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/abc/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_staff")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_shop_id_format", response["error"])
	})

	t.Run("when no shop_id in URL but shopID in context then checks permission", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		mockPermissionService.EXPECT().
			HasPermission("admin", "manage_products").
			Return(true)

		shopRoles := []claims.ShopRole{{ShopID: 5, Role: "admin"}}
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		// No shop_id in URL vars
		req = mux.SetURLVars(req, map[string]string{})
		rr := httptest.NewRecorder()

		// Act
		mw.RequirePermission("manage_products")(okHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when no shop_id in URL and no context shopID then returns 403", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		// Empty context — no shop roles, no shop IDs
		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		req = mux.SetURLVars(req, map[string]string{})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_products")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "insufficient_permissions", response["error"])
	})

	t.Run("when shop_id is zero then returns 400", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		shopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/0/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_staff")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_shop_id_format", response["error"])
	})

	t.Run("when shop_id is negative then returns 400", func(t *testing.T) {
		// Arrange
		mockPermissionService := mocks.NewPermissionService(t)
		mw := NewPermissionMiddleware(mockPermissionService)

		shopRoles := []claims.ShopRole{{ShopID: 1, Role: "owner"}}
		req := httptest.NewRequest(http.MethodGet, "/shops/-1/staff", nil)
		req = req.WithContext(contextWithShopRoles(shopRoles))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-1"})
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		// Act
		mw.RequirePermission("manage_staff")(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
