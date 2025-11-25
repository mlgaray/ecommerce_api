package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// mockIdentifiable is a test helper that implements both Identifiable and Sortable interfaces
type mockIdentifiable struct {
	ID   int
	Name string
}

func (m *mockIdentifiable) GetID() int {
	return m.ID
}

func (m *mockIdentifiable) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case "name":
		return m.Name
	case "id":
		return nil
	default:
		return nil
	}
}

func TestPaginationService_ApplyPagination(t *testing.T) {
	t.Run("when items equal limit then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange - LIMIT+1 strategy: if we got exactly 'limit' items, there are no more pages
		limit := 3
		products := []*models.Product{
			{ID: 1, Name: "Product 1"},
			{ID: 2, Name: "Product 2"},
			{ID: 3, Name: "Product 3"},
		}

		service := NewPaginationService[*models.Product]()

		// Act - sorting by ID
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 3, len(trimmedItems), "should return all items when count equals limit")
		assert.Empty(t, nextCursor, "nextCursor should be empty when no more pages")
		assert.False(t, hasMore, "hasMore should be false when items equal limit (LIMIT+1 strategy)")
	})

	t.Run("when items less than limit then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 5
		products := []*models.Product{
			{ID: 1, Name: "Product 1"},
			{ID: 2, Name: "Product 2"},
		}

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 2, len(trimmedItems), "should return all items when count less than limit")
		assert.Empty(t, nextCursor, "nextCursor should be empty when items less than limit")
		assert.False(t, hasMore, "hasMore should be false when items less than limit")
	})

	t.Run("when items list is empty then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 10
		products := []*models.Product{}

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 0, len(trimmedItems), "should return empty list when items list is empty")
		assert.Empty(t, nextCursor, "nextCursor should be empty when items list is empty")
		assert.False(t, hasMore, "hasMore should be false when items list is empty")
	})

	t.Run("when items list is nil then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 10
		var products []*models.Product

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Nil(t, trimmedItems, "should return nil when items list is nil")
		assert.Empty(t, nextCursor, "nextCursor should be empty when items list is nil")
		assert.False(t, hasMore, "hasMore should be false when items list is nil")
	})

	t.Run("when limit is 1 and has one item then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange - LIMIT+1 strategy: 1 item with limit=1 means no more pages
		limit := 1
		products := []*models.Product{
			{ID: 42, Name: "Single Product"},
		}

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 1, len(trimmedItems), "should return 1 item")
		assert.Empty(t, nextCursor, "nextCursor should be empty when no more pages")
		assert.False(t, hasMore, "hasMore should be false when items equal limit (LIMIT+1 strategy)")
	})

	t.Run("when using different identifiable type then works correctly", func(t *testing.T) {
		// Arrange - LIMIT+1 strategy: 3 items with limit=2 means there are more pages
		limit := 2
		items := []*mockIdentifiable{
			{ID: 10, Name: "Item 1"},
			{ID: 20, Name: "Item 2"},
			{ID: 30, Name: "Item 3"}, // Extra item indicates more pages
		}

		service := NewPaginationService[*mockIdentifiable]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(items, limit, "id")

		// Assert
		assert.Equal(t, 2, len(trimmedItems), "should return 2 items (trimmed)")
		assert.NotEmpty(t, nextCursor, "nextCursor should not be empty when there are more pages")
		assert.True(t, hasMore, "hasMore should be true when items exceed limit")
	})

	t.Run("when limit is zero and items list is empty then returns empty cursor and hasMore false", func(t *testing.T) {
		// Arrange
		limit := 0
		products := []*models.Product{}

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 0, len(trimmedItems), "should return empty list")
		assert.Empty(t, nextCursor, "nextCursor should be empty")
		assert.False(t, hasMore, "hasMore should be false")
	})

	t.Run("when items exceed limit by 1 then returns cursor and hasMore true", func(t *testing.T) {
		// Arrange - Simulating LIMIT+1 strategy
		limit := 3
		products := []*models.Product{
			{ID: 1, Name: "Product 1"},
			{ID: 2, Name: "Product 2"},
			{ID: 3, Name: "Product 3"},
			{ID: 4, Name: "Product 4"}, // Extra item from LIMIT+1
		}

		service := NewPaginationService[*models.Product]()

		// Act
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "id")

		// Assert
		assert.Equal(t, 3, len(trimmedItems), "should trim to limit (3 items)")
		assert.NotEmpty(t, nextCursor, "nextCursor should not be empty")
		assert.True(t, hasMore, "hasMore should be true when items exceed limit")
	})

	t.Run("when sorting by name field then includes sort value in cursor", func(t *testing.T) {
		// Arrange - LIMIT+1 strategy: 3 items with limit=2 means there are more pages
		limit := 2
		products := []*models.Product{
			{ID: 1, Name: "Product A"},
			{ID: 2, Name: "Product B"},
			{ID: 3, Name: "Product C"}, // Extra item indicates more pages
		}

		service := NewPaginationService[*models.Product]()

		// Act - sorting by name
		trimmedItems, nextCursor, hasMore := service.ApplyPagination(products, limit, "name")

		// Assert
		assert.Equal(t, 2, len(trimmedItems), "should return 2 items (trimmed)")
		assert.NotEmpty(t, nextCursor, "nextCursor should not be empty when there are more pages")
		assert.True(t, hasMore, "hasMore should be true when items exceed limit")

		// Verify cursor can be decoded and contains sort value
		// The cursor should point to the LAST VALID item (Product B, not the extra Product C)
		decoded, err := DecodeCursor(nextCursor)
		assert.NoError(t, err, "cursor should be decodable")
		assert.Equal(t, 2, decoded.ID, "decoded ID should match last valid item")
		assert.Equal(t, "Product B", decoded.SortValue, "decoded SortValue should match last valid item's name")
	})
}
