package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newTestFeaturedProducts creates test featured products for testing
func newTestFeaturedProducts() []*models.Product {
	return []*models.Product{
		{ID: 1, Name: "Featured Product 1", Price: 99.99, IsHighlighted: true, IsActive: true, CreatedAt: time.Now()},
		{ID: 2, Name: "Featured Product 2", Price: 149.99, IsHighlighted: true, IsActive: true, CreatedAt: time.Now()},
		{ID: 3, Name: "Featured Product 3", Price: 199.99, IsHighlighted: true, IsActive: true, CreatedAt: time.Now()},
	}
}

// =============================================================================
// GetStoreFeaturedProductsUseCase Tests
// =============================================================================

//nolint:gocyclo // Test functions with multiple sub-tests have high cyclomatic complexity by design
func TestGetStoreFeaturedProductsUseCase_Execute(t *testing.T) {
	t.Run("when store exists and has featured products then returns products", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		filters := models.ProductFilters{}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 20, "created_at").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("when store not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "non-existent-store"
		filters := models.ProductFilters{}
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
	})

	t.Run("when store exists but has no featured products then returns empty list", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "empty-store"
		store := newValidStore()
		emptyProducts := []*models.Product{}
		filters := models.ProductFilters{}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.ProductFilters")).
			Return(emptyProducts, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(emptyProducts, 20, "created_at").
			Return(emptyProducts, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("when product service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		filters := models.ProductFilters{}
		expectedError := errors.New("database error")

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.ProductFilters")).
			Return(nil, expectedError)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("when cursor provided then returns next page", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		lastID := 5
		filters := models.ProductFilters{
			LastID: &lastID,
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 20, "created_at").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("when has more pages then returns cursor", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		filters := models.ProductFilters{Limit: 2}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		expectedCursor := "eyJsYXN0X2lkIjoyLCJsYXN0X3ZhbHVlIjoiMjAyNC0wMS0xNVQxMDozMDowMFoifQ=="
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 2, "created_at").
			Return(products[:2], expectedCursor, true)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
	})

	t.Run("filters are forced to is_highlighted=true and is_active=true", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()

		// User tries to override filters - should be ignored
		isHighlighted := false
		isActive := false
		filters := models.ProductFilters{
			IsHighlighted: &isHighlighted,
			IsActive:      &isActive,
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		// Verify that filters passed have is_highlighted=true and is_active=true
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.IsHighlighted != nil && *f.IsHighlighted == true &&
					f.IsActive != nil && *f.IsActive == true
			})).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 20, "created_at").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, _, _, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
	})

	t.Run("when invalid sort field then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		filters := models.ProductFilters{
			SortBy: "invalid_field", // Invalid sort field
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		var validationErr *domainErrors.ValidationError
		assert.True(t, errors.As(err, &validationErr))
	})

	t.Run("when limit is zero then normalizes to default", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		filters := models.ProductFilters{
			Limit: 0, // Should be normalized to default (20)
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.Limit == 20 && // Default limit
					f.IsHighlighted != nil && *f.IsHighlighted == true &&
					f.IsActive != nil && *f.IsActive == true
			})).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 20, "created_at").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, _, _, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when limit exceeds max then normalizes to max", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		filters := models.ProductFilters{
			Limit: 1000, // Should be normalized to max (100)
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.Limit == 100 // Max limit
			})).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 100, "created_at").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, _, _, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("other user filters pass through with forced featured filters", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		products := newTestFeaturedProducts()
		categoryID := 5
		search := "phone"
		filters := models.ProductFilters{
			CategoryID: &categoryID,
			Search:     &search,
			SortBy:     "price",
			SortOrder:  "asc",
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		productServiceMock := mocks.NewProductService(t)
		// Verify that user filters pass through with forced featured filters
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.IsHighlighted != nil && *f.IsHighlighted == true &&
					f.IsActive != nil && *f.IsActive == true &&
					f.CategoryID != nil && *f.CategoryID == 5 &&
					f.Search != nil && *f.Search == "phone" &&
					f.SortBy == "price" &&
					f.SortOrder == "asc"
			})).
			Return(products, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(products, 20, "price").
			Return(products, "", false)

		useCase := NewGetStoreFeaturedProductsUseCase(storeServiceMock, productServiceMock, paginationServiceMock)

		// Act
		result, _, _, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
	})
}
