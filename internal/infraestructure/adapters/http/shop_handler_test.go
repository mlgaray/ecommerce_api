package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/auth/claims"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestShop() *models.Shop {
	primaryColor := "#FF5733"
	return &models.Shop{
		ID:           1,
		Name:         "Test Shop",
		Slug:         "test-shop",
		Email:        "test@shop.com",
		Phone:        "+54 11 1234-5678",
		Instagram:    "@testshop",
		PrimaryColor: &primaryColor,
		Images: []*models.Image{
			{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
		},
		Address: &models.Address{
			ID:   1,
			Name: "Main Street 123",
		},
	}
}

func contextWithShopIDs(shopIDs []int) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, claims.KeyShopIDs, shopIDs)
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestShopHandler_GetByID(t *testing.T) {
	t.Run("when shop exists and user owns shop then returns 200 with shop data", func(t *testing.T) {
		// Arrange
		shop := newTestShop()

		useCaseMock := mocks.NewGetShopByIDUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1).
			Return(shop, nil)

		handler := NewShopHandler(useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1", nil)
		req = req.WithContext(contextWithShopIDs([]int{1, 2, 3}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response models.Shop
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, shop.ID, response.ID)
		assert.Equal(t, shop.Name, response.Name)
		assert.Equal(t, shop.Slug, response.Slug)
	})

	t.Run("when shop not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetShopByIDUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 999).
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.ShopNotFound})

		handler := NewShopHandler(useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/999", nil)
		req = req.WithContext(contextWithShopIDs([]int{999}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when user does not own shop then returns 403", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1", nil)
		req = req.WithContext(contextWithShopIDs([]int{2, 3}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"}) // User owns shops 2 and 3, not 1
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("when shop_id is invalid format then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is zero then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/0", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is negative then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/-1", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when user has no shops then returns 403", func(t *testing.T) {
		// Arrange
		handler := NewShopHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1", nil)
		req = req.WithContext(contextWithShopIDs([]int{}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"}) // User owns no shops
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("when user owns single shop and requests it then returns 200", func(t *testing.T) {
		// Arrange
		shop := newTestShop()

		useCaseMock := mocks.NewGetShopByIDUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1).
			Return(shop, nil)

		handler := NewShopHandler(useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1", nil)
		req = req.WithContext(contextWithShopIDs([]int{1}))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"}) // User owns only shop 1
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// =============================================================================
// userOwnsShop Tests
// =============================================================================

func TestShopHandler_userOwnsShop(t *testing.T) {
	handler := &ShopHandler{}

	t.Run("when shop ID is in user shops then returns true", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop([]int{1, 2, 3}, 2)

		// Assert
		assert.True(t, result)
	})

	t.Run("when shop ID is not in user shops then returns false", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop([]int{1, 2, 3}, 5)

		// Assert
		assert.False(t, result)
	})

	t.Run("when user shops is empty then returns false", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop([]int{}, 1)

		// Assert
		assert.False(t, result)
	})

	t.Run("when user shops is nil then returns false", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop(nil, 1)

		// Assert
		assert.False(t, result)
	})

	t.Run("when shop ID is first in list then returns true", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop([]int{5, 10, 15}, 5)

		// Assert
		assert.True(t, result)
	})

	t.Run("when shop ID is last in list then returns true", func(t *testing.T) {
		// Act
		result := handler.userOwnsShop([]int{5, 10, 15}, 15)

		// Assert
		assert.True(t, result)
	})
}

// =============================================================================
// parseShopID Tests
// =============================================================================

func TestShopHandler_parseShopID(t *testing.T) {
	handler := &ShopHandler{}

	t.Run("when valid positive integer then returns shop ID", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/42", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "42"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 42, shopID)
	})

	t.Run("when zero then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/0", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when negative then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/-5", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-5"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when non-numeric then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when empty then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": ""})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when float then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/3.14", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "3.14"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when very large number then returns shop ID", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/999999999", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "999999999"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 999999999, shopID)
	})
}
