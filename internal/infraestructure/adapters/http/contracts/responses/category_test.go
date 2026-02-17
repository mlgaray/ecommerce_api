package responses

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestCategory() *models.Category {
	return &models.Category{
		ID:          1,
		Name:        "Electronics",
		Description: "Electronic products",
		Image: &models.Image{
			ID:  10,
			URL: "https://example.com/electronics.jpg",
		},
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
}

// =============================================================================
// CategoryResponseFromModel Tests
// =============================================================================

func TestCategoryResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := CategoryResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when category with image then maps all fields", func(t *testing.T) {
		// Arrange
		category := newTestCategory()

		// Act
		result := CategoryResponseFromModel(category)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Electronics", result.Name)
		assert.Equal(t, "Electronic products", result.Description)
		assert.Equal(t, time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), result.CreatedAt)
		assert.NotNil(t, result.Image)
		assert.Equal(t, 10, result.Image.ID)
		assert.Equal(t, "https://example.com/electronics.jpg", result.Image.URL)
	})

	t.Run("when category without image then image is nil", func(t *testing.T) {
		// Arrange
		category := &models.Category{
			ID:          2,
			Name:        "Clothing",
			Description: "Clothing products",
		}

		// Act
		result := CategoryResponseFromModel(category)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.ID)
		assert.Equal(t, "Clothing", result.Name)
		assert.Nil(t, result.Image)
	})

	t.Run("when category with zero values then maps zero values", func(t *testing.T) {
		// Arrange
		category := &models.Category{}

		// Act
		result := CategoryResponseFromModel(category)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ID)
		assert.Equal(t, "", result.Name)
		assert.Nil(t, result.Image)
	})
}

// =============================================================================
// CategoryResponsesFromModels Tests
// =============================================================================

func TestCategoryResponsesFromModels(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := CategoryResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when empty slice then returns empty slice", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{}

		// Act
		result := CategoryResponsesFromModels(categories)

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("when multiple categories then maps all", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Clothing", Image: &models.Image{ID: 5, URL: "https://example.com/img.jpg"}},
		}

		// Act
		result := CategoryResponsesFromModels(categories)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].ID)
		assert.Equal(t, "Electronics", result[0].Name)
		assert.Nil(t, result[0].Image)
		assert.Equal(t, 2, result[1].ID)
		assert.Equal(t, "Clothing", result[1].Name)
		assert.NotNil(t, result[1].Image)
	})
}

// =============================================================================
// NewListCategoriesResponse Tests
// =============================================================================

func TestNewListCategoriesResponse(t *testing.T) {
	t.Run("when categories with pagination then maps all fields", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{
			{ID: 1, Name: "Electronics"},
			{ID: 2, Name: "Clothing"},
		}
		totalCount := 50

		// Act
		result := NewListCategoriesResponse(categories, "cursor123", true, &totalCount)

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result.Categories, 2)
		assert.Equal(t, "cursor123", result.NextCursor)
		assert.True(t, result.HasMore)
		assert.NotNil(t, result.TotalCount)
		assert.Equal(t, 50, *result.TotalCount)
	})

	t.Run("when no more pages then has_more is false", func(t *testing.T) {
		// Arrange
		categories := []*models.Category{{ID: 1, Name: "Only"}}

		// Act
		result := NewListCategoriesResponse(categories, "", false, nil)

		// Assert
		assert.Len(t, result.Categories, 1)
		assert.Equal(t, "", result.NextCursor)
		assert.False(t, result.HasMore)
		assert.Nil(t, result.TotalCount)
	})

	t.Run("when nil categories then categories is nil", func(t *testing.T) {
		// Act
		result := NewListCategoriesResponse(nil, "", false, nil)

		// Assert
		assert.Nil(t, result.Categories)
		assert.False(t, result.HasMore)
	})
}
