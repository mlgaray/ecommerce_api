package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	authclaims "github.com/mlgaray/ecommerce_api/internal/core/claims"
	websocketHub "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/websocket"
	"github.com/mlgaray/ecommerce_api/mocks"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Connect Tests - Authentication
// =============================================================================

func TestOrderWSHandler_Connect_MissingToken(t *testing.T) {
	t.Run("when token is missing then returns 401", func(t *testing.T) {
		// Arrange
		hub := websocketHub.NewHub()
		tokenServiceMock := mocks.NewTokenService(t)
		handler := NewOrderWSHandler(hub, tokenServiceMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders/ws", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Connect(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestOrderWSHandler_Connect_InvalidToken(t *testing.T) {
	t.Run("when token is invalid then returns 401", func(t *testing.T) {
		// Arrange
		hub := websocketHub.NewHub()
		tokenServiceMock := mocks.NewTokenService(t)
		tokenServiceMock.EXPECT().
			ValidateAndParseClaims("invalid-token").
			Return(nil, &domainErrors.AuthenticationError{Message: "invalid_token"})

		handler := NewOrderWSHandler(hub, tokenServiceMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders/ws?token=invalid-token", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Connect(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

// =============================================================================
// Connect Tests - Authorization
// =============================================================================

func TestOrderWSHandler_Connect_Forbidden(t *testing.T) {
	t.Run("when user does not own shop then returns 403", func(t *testing.T) {
		// Arrange
		hub := websocketHub.NewHub()
		tokenServiceMock := mocks.NewTokenService(t)
		tokenServiceMock.EXPECT().
			ValidateAndParseClaims("valid-token").
			Return(&authclaims.TokenClaims{
				UserID:    1,
				Email:     "user@test.com",
				ShopRoles: []authclaims.ShopRole{{ShopID: 2, Role: "owner"}, {ShopID: 3, Role: "admin"}}, // Does NOT include shop 1
			}, nil)

		handler := NewOrderWSHandler(hub, tokenServiceMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/orders/ws?token=valid-token", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Connect(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

// =============================================================================
// Connect Tests - Invalid Shop ID
// =============================================================================

func TestOrderWSHandler_Connect_InvalidShopID(t *testing.T) {
	t.Run("when shop_id is not a number then returns 400", func(t *testing.T) {
		// Arrange
		hub := websocketHub.NewHub()
		tokenServiceMock := mocks.NewTokenService(t)
		tokenServiceMock.EXPECT().
			ValidateAndParseClaims("valid-token").
			Return(&authclaims.TokenClaims{
				UserID:    1,
				Email:     "user@test.com",
				ShopRoles: []authclaims.ShopRole{{ShopID: 1, Role: "owner"}},
			}, nil)

		handler := NewOrderWSHandler(hub, tokenServiceMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/orders/ws?token=valid-token", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.Connect(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is zero then returns 400", func(t *testing.T) {
		// Arrange
		hub := websocketHub.NewHub()
		tokenServiceMock := mocks.NewTokenService(t)
		tokenServiceMock.EXPECT().
			ValidateAndParseClaims("valid-token").
			Return(&authclaims.TokenClaims{
				UserID:    1,
				Email:     "user@test.com",
				ShopRoles: []authclaims.ShopRole{{ShopID: 1, Role: "owner"}},
			}, nil)

		handler := NewOrderWSHandler(hub, tokenServiceMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/0/orders/ws?token=valid-token", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})
		rr := httptest.NewRecorder()

		// Act
		handler.Connect(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestContainsShopID(t *testing.T) {
	t.Run("when shopID exists in slice then returns true", func(t *testing.T) {
		assert.True(t, containsShopID([]int{1, 2, 3}, 2))
	})

	t.Run("when shopID does not exist in slice then returns false", func(t *testing.T) {
		assert.False(t, containsShopID([]int{1, 2, 3}, 4))
	})

	t.Run("when slice is empty then returns false", func(t *testing.T) {
		assert.False(t, containsShopID([]int{}, 1))
	})

	t.Run("when slice is nil then returns false", func(t *testing.T) {
		assert.False(t, containsShopID(nil, 1))
	})
}
