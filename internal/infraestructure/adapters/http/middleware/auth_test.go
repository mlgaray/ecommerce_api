package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestMain(m *testing.M) {
	logs.Init()
	os.Exit(m.Run())
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	t.Run("when token is valid then injects claims into context and calls next handler", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		expectedClaims := &claims.TokenClaims{
			UserID:    1,
			Email:     "test@example.com",
			ShopRoles: []claims.ShopRole{{ShopID: 1, Role: "owner"}, {ShopID: 2, Role: "admin"}},
		}

		mockTokenService.EXPECT().
			ValidateAndParseClaims("valid-token").
			Return(expectedClaims, nil)

		var capturedUserID int
		var capturedEmail string
		var capturedShopIDs []int

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUserID = claims.GetUserIDFromContext(r.Context())
			capturedEmail = claims.GetEmailFromContext(r.Context())
			capturedShopIDs = claims.GetShopIDsFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, 1, capturedUserID)
		assert.Equal(t, "test@example.com", capturedEmail)
		assert.Equal(t, []int{1, 2}, capturedShopIDs)
	})

	t.Run("when authorization header is missing then returns 401", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "authorization_header_required", response["error"])
	})

	t.Run("when authorization header has no Bearer prefix then returns 401", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "some-token-without-bearer")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_token_format", response["error"])
	})

	t.Run("when token validation fails then returns error from token service", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		mockTokenService.EXPECT().
			ValidateAndParseClaims("expired-token").
			Return(nil, &errors.AuthenticationError{Message: errors.TokenExpired})

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, errors.TokenExpired, response["error"])
	})

	t.Run("when token is invalid then returns 401", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		mockTokenService.EXPECT().
			ValidateAndParseClaims("invalid-token").
			Return(nil, &errors.AuthenticationError{Message: errors.TokenInvalid})

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, errors.TokenInvalid, response["error"])
	})

	t.Run("when Bearer prefix has extra spaces then trims token correctly", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		expectedClaims := &claims.TokenClaims{
			UserID:    1,
			Email:     "test@example.com",
			ShopRoles: []claims.ShopRole{{ShopID: 1, Role: "owner"}},
		}

		mockTokenService.EXPECT().
			ValidateAndParseClaims("valid-token").
			Return(expectedClaims, nil)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer   valid-token  ")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when user has no shop_ids then injects empty slice", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)
		middleware := NewAuthMiddleware(mockTokenService)

		expectedClaims := &claims.TokenClaims{
			UserID:    1,
			Email:     "test@example.com",
			ShopRoles: []claims.ShopRole{},
		}

		mockTokenService.EXPECT().
			ValidateAndParseClaims(mock.Anything).
			Return(expectedClaims, nil)

		var capturedShopIDs []int

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedShopIDs = claims.GetShopIDsFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		// Act
		middleware.Authenticate(nextHandler).ServeHTTP(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, capturedShopIDs)
	})
}

func TestNewAuthMiddleware(t *testing.T) {
	t.Run("creates new auth middleware instance", func(t *testing.T) {
		// Arrange
		mockTokenService := mocks.NewTokenService(t)

		// Act
		middleware := NewAuthMiddleware(mockTokenService)

		// Assert
		assert.NotNil(t, middleware)
		assert.Equal(t, mockTokenService, middleware.tokenService)
	})
}
