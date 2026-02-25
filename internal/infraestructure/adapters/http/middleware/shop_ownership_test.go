package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
)

// passthrough handler that returns 200 OK
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func contextWithShopIDs(shopIDs []int) context.Context {
	return context.WithValue(context.Background(), claims.KeyShopIDs, shopIDs)
}

func TestShopOwnershipMiddleware_Authorize(t *testing.T) {
	mw := NewShopOwnershipMiddleware()

	t.Run("when user owns shop then passes through with 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{1, 2, 3}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when user does not own shop then returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{99}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("when user has empty shop IDs then returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("when user has nil shop IDs (no context) then returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders", nil)
		// No shop IDs in context at all
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("when shop_id is not numeric then returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/abc/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{1}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is zero then returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/0/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{1}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is negative then returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/shops/-1/orders", nil)
		req = req.WithContext(contextWithShopIDs([]int{1}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-1"})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id not present in URL then passes through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/some/other/path", nil)
		req = mux.SetURLVars(req, map[string]string{})
		rr := httptest.NewRecorder()

		mw.Authorize(okHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
