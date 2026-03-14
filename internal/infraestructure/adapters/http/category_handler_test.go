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

	"github.com/mlgaray/ecommerce_api/internal/core/claims"
	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestCategory() *models.Category {
	return &models.Category{
		ID:          1,
		Name:        "Electronics",
		Description: "Electronic devices",
		Image: &models.Image{
			ID:  1,
			URL: "https://example.com/electronics.jpg",
		},
	}
}

func categoryContextWithShopID(shopID int) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, claims.KeyShopIDs, []int{shopID})
	return ctx
}

// =============================================================================
// NewCategoryHandler Tests
// =============================================================================

func TestNewCategoryHandler(t *testing.T) {
	t.Run("creates handler with all dependencies", func(t *testing.T) {
		// Arrange
		createMock := mocks.NewCreateCategoryUseCase(t)
		updateMock := mocks.NewUpdateCategoryUseCase(t)
		deleteMock := mocks.NewDeleteCategoryUseCase(t)
		getByIDMock := mocks.NewGetCategoryByIDUseCase(t)
		getAllMock := mocks.NewGetAllCategoriesByShopIDWithFiltersUseCase(t)

		// Act
		handler := NewCategoryHandler(createMock, updateMock, deleteMock, getByIDMock, getAllMock)

		// Assert
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.createCategory)
		assert.NotNil(t, handler.updateCategory)
		assert.NotNil(t, handler.deleteCategory)
		assert.NotNil(t, handler.getByID)
		assert.NotNil(t, handler.getAllByShopIDWithFilters)
	})

	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		// Act
		handler := NewCategoryHandler(nil, nil, nil, nil, nil)

		// Assert
		assert.NotNil(t, handler)
	})
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestCategoryHandler_GetByID(t *testing.T) {
	t.Run("when category exists then returns 200 with category", func(t *testing.T) {
		// Arrange
		category := newTestCategory()

		getByIDMock := mocks.NewGetCategoryByIDUseCase(t)
		getByIDMock.EXPECT().
			Execute(mock.Anything, 1).
			Return(category, nil)

		handler := NewCategoryHandler(nil, nil, nil, getByIDMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories/1", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response responses.GetCategoryResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Category) {
			assert.Equal(t, category.ID, response.Category.ID)
			assert.Equal(t, category.Name, response.Category.Name)
		}
	})

	t.Run("when category not found then returns 404", func(t *testing.T) {
		// Arrange
		getByIDMock := mocks.NewGetCategoryByIDUseCase(t)
		getByIDMock.EXPECT().
			Execute(mock.Anything, 999).
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.CategoryNotFound})

		handler := NewCategoryHandler(nil, nil, nil, getByIDMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories/999", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when category_id invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewCategoryHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories/abc", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when category_id is zero then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewCategoryHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories/0", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "0"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetByID(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// GetAllByShopIDWithFilters Tests
// =============================================================================

func TestCategoryHandler_GetAllByShopIDWithFilters(t *testing.T) {
	t.Run("when categories exist then returns 200 with list", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Clothing"},
		}

		getAllMock := mocks.NewGetAllCategoriesByShopIDWithFiltersUseCase(t)
		getAllMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, "", false, (*int)(nil), nil)

		handler := NewCategoryHandler(nil, nil, nil, nil, getAllMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetAllByShopIDWithFilters(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["categories"])
	})

	t.Run("when shop not found then returns 404", func(t *testing.T) {
		// Arrange
		getAllMock := mocks.NewGetAllCategoriesByShopIDWithFiltersUseCase(t)
		getAllMock.EXPECT().
			Execute(mock.Anything, 999, mock.AnythingOfType("models.CategoryFilters")).
			Return(nil, "", false, (*int)(nil), &domainErrors.RecordNotFoundError{Message: domainErrors.ShopNotFound})

		handler := NewCategoryHandler(nil, nil, nil, nil, getAllMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/999/categories", nil)
		req = req.WithContext(categoryContextWithShopID(999))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetAllByShopIDWithFilters(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when shop_id invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewCategoryHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetAllByShopIDWithFilters(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when pagination has more then includes cursor", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Clothing"},
		}
		expectedCursor := "eyJsYXN0X2lkIjoyfQ=="

		getAllMock := mocks.NewGetAllCategoriesByShopIDWithFiltersUseCase(t)
		getAllMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, expectedCursor, true, (*int)(nil), nil)

		handler := NewCategoryHandler(nil, nil, nil, nil, getAllMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories?limit=2", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetAllByShopIDWithFilters(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCursor, response["next_cursor"])
		assert.Equal(t, true, response["has_more"])
	})

	t.Run("when search filter provided then passes to use case", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{}

		getAllMock := mocks.NewGetAllCategoriesByShopIDWithFiltersUseCase(t)
		getAllMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, "", false, (*int)(nil), nil)

		handler := NewCategoryHandler(nil, nil, nil, nil, getAllMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/categories?search=electronics", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetAllByShopIDWithFilters(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestCategoryHandler_Delete(t *testing.T) {
	t.Run("when category deleted successfully then returns 204", func(t *testing.T) {
		// Arrange
		deleteMock := mocks.NewDeleteCategoryUseCase(t)
		deleteMock.EXPECT().
			Execute(mock.Anything, 1, 1).
			Return(nil)

		handler := NewCategoryHandler(nil, nil, deleteMock, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/shops/1/categories/1", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Delete(rr, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("when category not found then returns 404", func(t *testing.T) {
		// Arrange
		deleteMock := mocks.NewDeleteCategoryUseCase(t)
		deleteMock.EXPECT().
			Execute(mock.Anything, 999, 1). // categoryID=999, shopID=1
			Return(&domainErrors.RecordNotFoundError{Message: domainErrors.CategoryNotFound})

		handler := NewCategoryHandler(nil, nil, deleteMock, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/shops/1/categories/999", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.Delete(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when category has products then returns 409", func(t *testing.T) {
		// Arrange
		deleteMock := mocks.NewDeleteCategoryUseCase(t)
		deleteMock.EXPECT().
			Execute(mock.Anything, 1, 1).
			Return(&domainErrors.ReferentialIntegrityError{Message: domainErrors.CategoryHasProducts})

		handler := NewCategoryHandler(nil, nil, deleteMock, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/shops/1/categories/1", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Delete(rr, req)

		// Assert
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("when category_id invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewCategoryHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/shops/1/categories/abc", nil)
		req = req.WithContext(categoryContextWithShopID(1))
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "category_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.Delete(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// parseCategoryID Tests
// =============================================================================

func TestCategoryHandler_parseCategoryID(t *testing.T) {
	handler := &CategoryHandler{}

	t.Run("when valid positive integer then returns ID", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/categories/42", nil)
		req = mux.SetURLVars(req, map[string]string{"category_id": "42"})

		// Act
		categoryID, err := handler.parseCategoryID(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 42, categoryID)
	})

	t.Run("when zero then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/categories/0", nil)
		req = mux.SetURLVars(req, map[string]string{"category_id": "0"})

		// Act
		categoryID, err := handler.parseCategoryID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, categoryID)
	})

	t.Run("when negative then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/categories/-5", nil)
		req = mux.SetURLVars(req, map[string]string{"category_id": "-5"})

		// Act
		categoryID, err := handler.parseCategoryID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, categoryID)
	})

	t.Run("when non-numeric then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/categories/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"category_id": "abc"})

		// Act
		categoryID, err := handler.parseCategoryID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, categoryID)
	})
}

// =============================================================================
// parseShopID Tests
// =============================================================================

func TestCategoryHandler_parseShopID(t *testing.T) {
	handler := &CategoryHandler{}

	t.Run("when valid positive integer then returns ID", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/42/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "42"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 42, shopID)
	})

	t.Run("when zero then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/0/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when negative then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/-5/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-5"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})

	t.Run("when non-numeric then returns error", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodGet, "/shops/abc/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})

		// Act
		shopID, err := handler.parseShopID(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, shopID)
	})
}
