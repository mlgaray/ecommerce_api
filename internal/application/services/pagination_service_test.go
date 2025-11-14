package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// mockIdentifiable is a test helper that implements Identifiable interface
type mockIdentifiable struct {
	ID   int
	Name string
}

func (m *mockIdentifiable) GetID() int {
	return m.ID
}

func TestPaginationService_BuildCursorPagination(t *testing.T) {
	t.Run("when items equal limit then returns cursor and hasMore true", func(t *testing.T) {
		// Arrange
		limit := 3
		products := []*models.Product{
			{ID: 1, Name: "Product 1"},
			{ID: 2, Name: "Product 2"},
			{ID: 3, Name: "Product 3"},
		}

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 3, nextCursor, "nextCursor should be the ID of the last item")
		assert.True(t, hasMore, "hasMore should be true when items equal limit")
	})

	t.Run("when items less than limit then returns zero cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 5
		products := []*models.Product{
			{ID: 1, Name: "Product 1"},
			{ID: 2, Name: "Product 2"},
		}

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 0, nextCursor, "nextCursor should be 0 when items less than limit")
		assert.False(t, hasMore, "hasMore should be false when items less than limit")
	})

	t.Run("when items list is empty then returns zero cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 10
		products := []*models.Product{}

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 0, nextCursor, "nextCursor should be 0 when items list is empty")
		assert.False(t, hasMore, "hasMore should be false when items list is empty")
	})

	t.Run("when items list is nil then returns zero cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 10
		var products []*models.Product

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 0, nextCursor, "nextCursor should be 0 when items list is nil")
		assert.False(t, hasMore, "hasMore should be false when items list is nil")
	})

	t.Run("when limit is 1 and has one item then returns cursor and hasMore true", func(t *testing.T) {
		// Arrange
		limit := 1
		products := []*models.Product{
			{ID: 42, Name: "Single Product"},
		}

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 42, nextCursor, "nextCursor should be the ID of the last item")
		assert.True(t, hasMore, "hasMore should be true when items equal limit")
	})

	t.Run("when using different identifiable type then works correctly", func(t *testing.T) {
		// Arrange
		limit := 2
		items := []*mockIdentifiable{
			{ID: 10, Name: "Item 1"},
			{ID: 20, Name: "Item 2"},
		}

		service := NewPaginationService[*mockIdentifiable]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(items, limit)

		// Assert
		assert.Equal(t, 20, nextCursor, "nextCursor should be the ID of the last item")
		assert.True(t, hasMore, "hasMore should be true when items equal limit")
	})

	t.Run("when limit is zero and items list is empty then returns zero cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 0
		products := []*models.Product{}

		service := NewPaginationService[*models.Product]()

		// Act
		nextCursor, hasMore := service.BuildCursorPagination(products, limit)

		// Assert
		assert.Equal(t, 0, nextCursor, "nextCursor should be 0")
		assert.False(t, hasMore, "hasMore should be false")
	})
}
