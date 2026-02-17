package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func init() {
	logs.Init()
}

// =============================================================================
// Test Helpers
// =============================================================================

func newTestStore() *models.Store {
	return &models.Store{
		ID:        1,
		Name:      "Test Store",
		Slug:      "test-store",
		Email:     "test@store.com",
		Phone:     "+54 11 1234-5678",
		Instagram: "@teststore",
		Images: []*models.Image{
			{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
		},
		Address: &models.Address{
			ID:   1,
			Name: "Main Street 123",
		},
	}
}

func newTestProductList() []*models.Product {
	return []*models.Product{
		{ID: 1, Name: "Product 1", Price: 99.99, IsActive: true, CreatedAt: time.Now()},
		{ID: 2, Name: "Product 2", Price: 149.99, IsActive: true, CreatedAt: time.Now()},
	}
}

func newTestCategoryList() []*models.Category {
	return []*models.Category{
		{ID: 1, Name: "Electronics"},
		{ID: 2, Name: "Clothing"},
	}
}

// =============================================================================
// GetBySlug Tests
// =============================================================================

func TestStoreHandler_GetBySlug(t *testing.T) {
	t.Run("when store exists then returns 200 with store data", func(t *testing.T) {
		// Arrange
		store := newTestStore()

		useCaseMock := mocks.NewGetStoreBySlugUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store").
			Return(store, nil)

		handler := NewStoreHandler(useCaseMock, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetBySlug(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response responses.GetStoreResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Store) {
			assert.Equal(t, store.ID, response.Store.ID)
			assert.Equal(t, store.Name, response.Store.Name)
			assert.Equal(t, store.Slug, response.Store.Slug)
		}
	})

	t.Run("when store not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreBySlugUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "non-existent").
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound})

		handler := NewStoreHandler(useCaseMock, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/non-existent", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "non-existent"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetBySlug(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewStoreHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.GetBySlug(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when slug is whitespace only then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewStoreHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/whitespace-test", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "   "})
		rr := httptest.NewRecorder()

		// Act
		handler.GetBySlug(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// GetCategories Tests
// =============================================================================

func TestStoreHandler_GetCategories(t *testing.T) {
	t.Run("when store exists and has categories then returns 200", func(t *testing.T) {
		// Arrange
		categories := newTestCategoryList()

		useCaseMock := mocks.NewGetStoreCategoriesUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, "", false, nil)

		handler := NewStoreHandler(nil, useCaseMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetCategories(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["categories"])
	})

	t.Run("when store not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreCategoriesUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "non-existent", mock.AnythingOfType("models.CategoryFilters")).
			Return(nil, "", false, &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound})

		handler := NewStoreHandler(nil, useCaseMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/non-existent/categories", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "non-existent"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetCategories(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewStoreHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores//categories", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.GetCategories(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when search filter provided then passes to use case", func(t *testing.T) {
		// Arrange
		categories := newTestCategoryList()

		useCaseMock := mocks.NewGetStoreCategoriesUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, "", false, nil)

		handler := NewStoreHandler(nil, useCaseMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/categories?search=electronics", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetCategories(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when pagination has more then includes cursor", func(t *testing.T) {
		// Arrange
		categories := newTestCategoryList()
		expectedCursor := "eyJsYXN0X2lkIjoyLCJsYXN0X3ZhbHVlIjoiQ2xvdGhpbmcifQ=="

		useCaseMock := mocks.NewGetStoreCategoriesUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, expectedCursor, true, nil)

		handler := NewStoreHandler(nil, useCaseMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/categories?limit=2", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetCategories(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCursor, response["next_cursor"])
		assert.Equal(t, true, response["has_more"])
	})
}

// =============================================================================
// GetProducts Tests
// =============================================================================

func TestStoreHandler_GetProducts(t *testing.T) {
	t.Run("when store exists and has products then returns 200", func(t *testing.T) {
		// Arrange
		products := newTestProductList()

		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, "", false, nil)

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["products"])
	})

	t.Run("when store not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "non-existent", mock.AnythingOfType("models.ProductFilters")).
			Return(nil, "", false, &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound})

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/non-existent/products", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "non-existent"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewStoreHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores//products", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when search filter provided then passes to use case", func(t *testing.T) {
		// Arrange
		products := newTestProductList()

		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, "", false, nil)

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products?search=phone", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when category filter provided then passes to use case", func(t *testing.T) {
		// Arrange
		products := newTestProductList()

		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, "", false, nil)

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products?category_id=5", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when pagination has more then includes cursor", func(t *testing.T) {
		// Arrange
		products := newTestProductList()
		expectedCursor := "eyJsYXN0X2lkIjoyLCJsYXN0X3ZhbHVlIjoiMjAyNC0wMS0xNVQxMDozMDowMFoifQ=="

		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, expectedCursor, true, nil)

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products?limit=2", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCursor, response["next_cursor"])
		assert.Equal(t, true, response["has_more"])
	})

	t.Run("when validation error then returns 400", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(nil, "", false, &domainErrors.ValidationError{Message: domainErrors.InvalidSortField})

		handler := NewStoreHandler(nil, nil, useCaseMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// GetFeaturedProducts Tests
// =============================================================================

func TestStoreHandler_GetFeaturedProducts(t *testing.T) {
	t.Run("when store exists and has featured products then returns 200", func(t *testing.T) {
		// Arrange
		products := newTestProductList()

		useCaseMock := mocks.NewGetStoreFeaturedProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, "", false, nil)

		handler := NewStoreHandler(nil, nil, nil, useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products/featured", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["products"])
	})

	t.Run("when store not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreFeaturedProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "non-existent", mock.AnythingOfType("models.ProductFilters")).
			Return(nil, "", false, &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound})

		handler := NewStoreHandler(nil, nil, nil, useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/non-existent/products/featured", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "non-existent"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewStoreHandler(nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores//products/featured", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when pagination has more then includes cursor", func(t *testing.T) {
		// Arrange
		products := newTestProductList()
		expectedCursor := "eyJsYXN0X2lkIjoyLCJsYXN0X3ZhbHVlIjoiMjAyNC0wMS0xNVQxMDozMDowMFoifQ=="

		useCaseMock := mocks.NewGetStoreFeaturedProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(products, expectedCursor, true, nil)

		handler := NewStoreHandler(nil, nil, nil, useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products/featured?limit=2", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCursor, response["next_cursor"])
		assert.Equal(t, true, response["has_more"])
	})

	t.Run("when validation error then returns 400", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreFeaturedProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return(nil, "", false, &domainErrors.ValidationError{Message: domainErrors.InvalidSortField})

		handler := NewStoreHandler(nil, nil, nil, useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products/featured", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when empty result then returns empty array", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewGetStoreFeaturedProductsUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, "test-store", mock.AnythingOfType("models.ProductFilters")).
			Return([]*models.Product{}, "", false, nil)

		handler := NewStoreHandler(nil, nil, nil, useCaseMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/stores/test-store/products/featured", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetFeaturedProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		products, ok := response["products"].([]interface{})
		assert.True(t, ok, "products should be an array")
		assert.Empty(t, products)
	})
}
