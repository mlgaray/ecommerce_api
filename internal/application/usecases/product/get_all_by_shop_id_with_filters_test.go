package product

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
	t.Run("when first page request then returns products with total count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Product 1", CreatedAt: time.Now()},
			{ID: 2, Name: "Product 2", CreatedAt: time.Now()},
		}
		expectedCount := 50

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(expectedCount, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, products, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when first page with more pages then returns cursor and hasMore true", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.ProductFilters{
			Limit:     2,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		// Repository returns 3 items (limit+1 strategy)
		productsFromRepo := []*models.Product{
			{ID: 1, Name: "Product 1", CreatedAt: time.Now()},
			{ID: 2, Name: "Product 2", CreatedAt: time.Now()},
			{ID: 3, Name: "Product 3", CreatedAt: time.Now()},
		}
		// Pagination service trims to 2
		trimmedProducts := productsFromRepo[:2]
		expectedCursor := "eyJpZCI6MiwidjoiMjAyNS0wMS0wMVQxMDowMDowMCJ9"
		expectedCount := 100

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(expectedCount, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(productsFromRepo, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(productsFromRepo, filters.Limit, filters.SortBy).
			Return(trimmedProducts, expectedCursor, true)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, trimmedProducts, result)
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
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
			LastID:    &lastID,
		}

		products := []*models.Product{
			{ID: 11, Name: "Product 11", CreatedAt: time.Now()},
			{ID: 12, Name: "Product 12", CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		// CountByShopIDWithFilters should NOT be called on subsequent pages
		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, products, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount) // No total count on subsequent pages
	})

	t.Run("when product service returns error then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
		expectedError := errors.New("database error")

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(0, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(nil, expectedError)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

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

	t.Run("when count fails then still returns products without total count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Product 1", CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		// Count fails but should not fail the request
		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(0, errors.New("count error"))

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, products, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount) // Count failed, so nil
	})

	t.Run("when empty result then returns empty list", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 999
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		emptyProducts := []*models.Product{}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(0, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(emptyProducts, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(emptyProducts, filters.Limit, filters.SortBy).
			Return(emptyProducts, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

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
		searchTerm := "Laptop"
		filters := models.ProductFilters{
			Search:    &searchTerm,
			Limit:     20,
			SortBy:    "name",
			SortOrder: "asc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Laptop Pro", CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.Search != nil && *f.Search == searchTerm
			})).
			Return(1, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, _, _, _, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Laptop Pro", result[0].Name)
	})

	t.Run("when category filter applied then passes filter to service", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		categoryID := 5
		filters := models.ProductFilters{
			CategoryID: &categoryID,
			Limit:      20,
			SortBy:     "created_at",
			SortOrder:  "desc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Product in Category 5", CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.CategoryID != nil && *f.CategoryID == categoryID
			})).
			Return(1, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, _, _, _, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("when price range filter applied then passes filter to service", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		minPrice := 100.0
		maxPrice := 500.0
		filters := models.ProductFilters{
			MinPrice:  &minPrice,
			MaxPrice:  &maxPrice,
			Limit:     20,
			SortBy:    "price",
			SortOrder: "asc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Mid-range Product", Price: 250.0, CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.MinPrice != nil && *f.MinPrice == minPrice &&
					f.MaxPrice != nil && *f.MaxPrice == maxPrice
			})).
			Return(1, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, _, _, _, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("when isActive filter applied then passes filter to service", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		isActive := true
		filters := models.ProductFilters{
			IsActive:  &isActive,
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		products := []*models.Product{
			{ID: 1, Name: "Active Product", IsActive: true, CreatedAt: time.Now()},
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		productServiceMock.EXPECT().
			CountByShopIDWithFilters(ctx, shopID, mock.MatchedBy(func(f models.ProductFilters) bool {
				return f.IsActive != nil && *f.IsActive == isActive
			})).
			Return(1, nil)

		productServiceMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, shopID, mock.AnythingOfType("models.ProductFilters")).
			Return(products, nil)

		paginationServiceMock.EXPECT().
			ApplyPagination(products, filters.Limit, filters.SortBy).
			Return(products, "", false)

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, _, _, _, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("when validation fails then returns error without calling services", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		invalidMinPrice := -10.0
		filters := models.ProductFilters{
			MinPrice:  &invalidMinPrice, // Negative price should fail validation
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		// Services should NOT be called when validation fails

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when invalid sort field then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.ProductFilters{
			Limit:     20,
			SortBy:    "invalid_field",
			SortOrder: "desc",
		}

		productServiceMock := mocks.NewProductService(t)
		paginationServiceMock := mocks.NewPaginationService[*models.Product](t)

		// Services should NOT be called when validation fails

		useCase := NewGetAllByShopIDWithFiltersUseCase(productServiceMock, paginationServiceMock)

		// Act
		result, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}
