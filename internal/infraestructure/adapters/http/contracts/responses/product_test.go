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

func newTestProduct() *models.Product {
	return &models.Product{
		ID:               1,
		Name:             "Laptop",
		Description:      "A powerful laptop",
		Price:            999.99,
		IsActive:         true,
		IsPromotional:    true,
		PromotionalPrice: 799.99,
		IsHighlighted:    true,
		IsStockeable:     true,
		Stock:            50,
		MinimumStock:     5,
		CreatedAt:        time.Date(2024, 3, 10, 14, 0, 0, 0, time.UTC),
		Images: []*models.Image{
			{ID: 10, URL: "https://example.com/laptop1.jpg"},
			{ID: 11, URL: "https://example.com/laptop2.jpg"},
		},
		Category: &models.Category{
			ID:          5,
			Name:        "Electronics",
			Description: "Electronic products",
		},
		Variants: []*models.Variant{
			{
				ID:            20,
				Name:          "Color",
				Order:         1,
				SelectionType: models.Single,
				MaxSelections: 1,
				IsRequired:    true,
				Options: []*models.Option{
					{ID: 100, Name: "Silver", Price: 0, Order: 1},
					{ID: 101, Name: "Gold", Price: 50, Order: 2},
				},
			},
		},
	}
}

// =============================================================================
// ProductResponseFromModel Tests
// =============================================================================

func TestProductResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ProductResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when full product then maps all fields", func(t *testing.T) {
		// Arrange
		product := newTestProduct()

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Laptop", result.Name)
		assert.Equal(t, "A powerful laptop", result.Description)
		assert.Equal(t, 999.99, result.Price)
		assert.True(t, result.IsActive)
		assert.True(t, result.IsPromotional)
		assert.Equal(t, 799.99, result.PromotionalPrice)
		assert.True(t, result.IsHighlighted)
		assert.True(t, result.IsStockeable)
		assert.Equal(t, 50, result.Stock)
		assert.Equal(t, 5, result.MinimumStock)
		assert.Equal(t, time.Date(2024, 3, 10, 14, 0, 0, 0, time.UTC), result.CreatedAt)
	})

	t.Run("when product has images then maps images", func(t *testing.T) {
		// Arrange
		product := newTestProduct()

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Len(t, result.Images, 2)
		assert.Equal(t, 10, result.Images[0].ID)
		assert.Equal(t, "https://example.com/laptop1.jpg", result.Images[0].URL)
		assert.Equal(t, 11, result.Images[1].ID)
	})

	t.Run("when product has no images then images is nil", func(t *testing.T) {
		// Arrange
		product := &models.Product{ID: 1, Name: "Simple"}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Nil(t, result.Images)
	})

	t.Run("when product has category then maps category", func(t *testing.T) {
		// Arrange
		product := newTestProduct()

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.NotNil(t, result.Category)
		assert.Equal(t, 5, result.Category.ID)
		assert.Equal(t, "Electronics", result.Category.Name)
		assert.Equal(t, "Electronic products", result.Category.Description)
	})

	t.Run("when product has no category then category is nil", func(t *testing.T) {
		// Arrange
		product := &models.Product{ID: 1, Name: "No Category"}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Nil(t, result.Category)
	})

	t.Run("when product has variants then maps variants with options", func(t *testing.T) {
		// Arrange
		product := newTestProduct()

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Len(t, result.Variants, 1)
		assert.Equal(t, 20, result.Variants[0].ID)
		assert.Equal(t, "Color", result.Variants[0].Name)
		assert.Equal(t, 1, result.Variants[0].Order)
		assert.Equal(t, "single", result.Variants[0].SelectionType)
		assert.Equal(t, 1, result.Variants[0].MaxSelections)
		assert.True(t, result.Variants[0].IsRequired)
		assert.Len(t, result.Variants[0].Options, 2)
		assert.Equal(t, 100, result.Variants[0].Options[0].ID)
		assert.Equal(t, "Silver", result.Variants[0].Options[0].Name)
		assert.Equal(t, 0.0, result.Variants[0].Options[0].Price)
		assert.Equal(t, 1, result.Variants[0].Options[0].Order)
	})

	t.Run("when product has no variants then variants is nil", func(t *testing.T) {
		// Arrange
		product := &models.Product{ID: 1, Name: "No Variants"}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Nil(t, result.Variants)
	})

	t.Run("when product has zero values then maps zero values", func(t *testing.T) {
		// Arrange
		product := &models.Product{}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ID)
		assert.False(t, result.IsActive)
		assert.Equal(t, 0.0, result.Price)
		assert.Equal(t, 0, result.Stock)
	})
}

// =============================================================================
// variantResponseFromModel Tests (via ProductResponseFromModel)
// =============================================================================

func TestVariantResponseFromModel(t *testing.T) {
	t.Run("when variant has no options then options is nil", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:   1,
			Name: "Test",
			Variants: []*models.Variant{
				{ID: 1, Name: "Size", SelectionType: models.Single},
			},
		}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Len(t, result.Variants, 1)
		assert.Nil(t, result.Variants[0].Options)
	})

	t.Run("when variant has multiple selection type then maps correctly", func(t *testing.T) {
		// Arrange
		product := &models.Product{
			ID:   1,
			Name: "Test",
			Variants: []*models.Variant{
				{
					ID:            1,
					Name:          "Toppings",
					SelectionType: models.Multiple,
					MaxSelections: 3,
					Options: []*models.Option{
						{ID: 1, Name: "Cheese", Price: 1.50, Order: 1},
					},
				},
			},
		}

		// Act
		result := ProductResponseFromModel(product)

		// Assert
		assert.Equal(t, "multiple", result.Variants[0].SelectionType)
		assert.Equal(t, 3, result.Variants[0].MaxSelections)
	})
}

// =============================================================================
// ProductResponsesFromModels Tests
// =============================================================================

func TestProductResponsesFromModels(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := ProductResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when empty slice then returns empty slice", func(t *testing.T) {
		// Act
		result := ProductResponsesFromModels([]*models.Product{})

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("when multiple products then maps all", func(t *testing.T) {
		// Arrange
		products := []*models.Product{
			{ID: 1, Name: "Product A", Price: 10.0},
			{ID: 2, Name: "Product B", Price: 20.0},
		}

		// Act
		result := ProductResponsesFromModels(products)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "Product A", result[0].Name)
		assert.Equal(t, "Product B", result[1].Name)
	})
}

// =============================================================================
// NewListProductsResponse Tests
// =============================================================================

func TestNewListProductsResponse(t *testing.T) {
	t.Run("when products with pagination then maps all fields", func(t *testing.T) {
		// Arrange
		products := []*models.Product{
			{ID: 1, Name: "Product A"},
			{ID: 2, Name: "Product B"},
		}
		totalCount := 100

		// Act
		result := NewListProductsResponse(products, "nextCur", true, &totalCount)

		// Assert
		assert.Len(t, result.Products, 2)
		assert.Equal(t, "nextCur", result.NextCursor)
		assert.True(t, result.HasMore)
		assert.Equal(t, 100, *result.TotalCount)
	})

	t.Run("when nil products then products is nil", func(t *testing.T) {
		// Act
		result := NewListProductsResponse(nil, "", false, nil)

		// Assert
		assert.Nil(t, result.Products)
		assert.False(t, result.HasMore)
		assert.Nil(t, result.TotalCount)
	})
}
