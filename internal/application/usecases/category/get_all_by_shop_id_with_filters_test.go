package category

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

func TestGetAllByShopIDWithFiltersUseCase_Execute(t *testing.T) {
	t.Run("when first page request then returns categories with total count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := &models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		categories := []*models.Category{
			{ID: 1, Name: "Electronics", CreatedAt: time.Now()},
			{ID: 2, Name: "Clothing", CreatedAt: time.Now()},
		}
		expectedCount := 50

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, *filters).
			Return(expectedCount, nil)

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(categories, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(categories, filters.Limit, filters.SortBy).
			Return(categories, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, categories, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when first page with more pages then returns cursor and hasMore true", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := &models.CategoryFilters{
			Limit:     2,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		// Repository returns 3 items (limit+1 strategy)
		categoriesFromRepo := []*models.Category{
			{ID: 1, Name: "Electronics", CreatedAt: time.Now()},
			{ID: 2, Name: "Clothing", CreatedAt: time.Now()},
			{ID: 3, Name: "Food", CreatedAt: time.Now()},
		}
		// Pagination service trims to 2
		trimmedCategories := categoriesFromRepo[:2]
		expectedCursor := "eyJpZCI6MiwidjoiMjAyNS0wMS0wMVQxMDowMDowMCJ9"
		expectedCount := 100

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, *filters).
			Return(expectedCount, nil)

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(categoriesFromRepo, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(categoriesFromRepo, filters.Limit, filters.SortBy).
			Return(trimmedCategories, expectedCursor, true)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, trimmedCategories, result)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
		assert.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when subsequent page then does not call count and returns nil totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		lastID := 10
		filters := &models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
			LastID:    &lastID,
		}

		categories := []*models.Category{
			{ID: 11, Name: "Books", CreatedAt: time.Now()},
			{ID: 12, Name: "Toys", CreatedAt: time.Now()},
		}

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		// CountByShopIDWithFilters should NOT be called on subsequent pages
		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(categories, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(categories, filters.Limit, filters.SortBy).
			Return(categories, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, categories, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount) // No total count on subsequent pages
	})

	t.Run("when category service returns error then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := &models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
		expectedError := errors.New("database error")

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, *filters).
			Return(0, nil)

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(nil, expectedError)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when count fails then still returns categories without total count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := &models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		categories := []*models.Category{
			{ID: 1, Name: "Electronics", CreatedAt: time.Now()},
		}

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		// Count fails but should not fail the request
		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, *filters).
			Return(0, errors.New("count error"))

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(categories, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(categories, filters.Limit, filters.SortBy).
			Return(categories, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, categories, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount) // Count failed, so nil
	})

	t.Run("when empty result then returns empty list", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		filters := &models.CategoryFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		emptyCategories := []*models.Category{}

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, *filters).
			Return(0, nil)

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(emptyCategories, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(emptyCategories, filters.Limit, filters.SortBy).
			Return(emptyCategories, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.NotNil(t, totalCount)
		assert.Equal(t, 0, *totalCount)
	})

	t.Run("when search filter applied then passes filter to service", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		searchTerm := "Electronics"
		filters := &models.CategoryFilters{
			Search:    &searchTerm,
			Limit:     20,
			SortBy:    "name",
			SortOrder: "asc",
		}

		categories := []*models.Category{
			{ID: 1, Name: "Electronics", CreatedAt: time.Now()},
		}

		categoryServiceMock := mocks.NewCategoryService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Category](t)

		categoryServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.CategoryFilters) bool {
				return f.Search != nil && *f.Search == searchTerm
			})).
			Return(1, nil)

		categoryServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, filters).
			Return(categories, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(categories, filters.Limit, filters.SortBy).
			Return(categories, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(categoryServiceMock, paginationServiceMock)

		// Act
		result, _, _, _, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Electronics", result[0].Name)
	})
}
