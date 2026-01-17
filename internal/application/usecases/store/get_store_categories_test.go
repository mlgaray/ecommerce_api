package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newTestCategories creates test categories for testing
func newTestCategories() []*models.Category {
	return []*models.Category{
		{ID: 1, Name: "Electronics"},
		{ID: 2, Name: "Clothing"},
		{ID: 3, Name: "Food"},
	}
}

// =============================================================================
// GetStoreCategoriesUseCase Tests
// =============================================================================

func TestGetStoreCategoriesUseCase_Execute(t *testing.T) {
	t.Run("when store exists and has categories then returns categories", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		categories := newTestCategories()
		filters := models.CategoryFilters{}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(categories, 20, "created_at").
			Return(categories, "", false)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

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
		filters := models.CategoryFilters{}
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

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

	t.Run("when store exists but has no categories then returns empty list", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "empty-store"
		store := newValidStore()
		emptyCategories := []*models.Category{}
		filters := models.CategoryFilters{}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.CategoryFilters")).
			Return(emptyCategories, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(emptyCategories, 20, "created_at").
			Return(emptyCategories, "", false)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
	})

	t.Run("when category service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		filters := models.CategoryFilters{}
		expectedError := errors.New("database error")

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.CategoryFilters")).
			Return(nil, expectedError)

		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

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
		categories := newTestCategories()
		lastID := 5
		filters := models.CategoryFilters{
			LastID: &lastID,
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, nil)

		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(categories, 20, "created_at").
			Return(categories, "", false)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

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
		categories := newTestCategories()
		filters := models.CategoryFilters{Limit: 2}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, store.ID, mock.AnythingOfType("models.CategoryFilters")).
			Return(categories, nil)

		expectedCursor := "eyJsYXN0X2lkIjoyLCJsYXN0X3ZhbHVlIjoiQ2xvdGhpbmcifQ=="
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)
		paginationServiceMock.EXPECT().
			ApplyPagination(categories, 2, "created_at").
			Return(categories[:2], expectedCursor, true)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, err := useCase.Execute(ctx, slug, filters)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
	})

	t.Run("when invalid filters then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		store := newValidStore()
		filters := models.CategoryFilters{
			SortBy: "invalid_field", // Invalid sort field
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(store, nil)

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		useCase := NewGetStoreCategoriesUseCase(storeServiceMock, categoryServiceMock, paginationServiceMock)

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
}
