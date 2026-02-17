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

func newTestOrder() *models.Order {
	return &models.Order{
		ID:          1,
		OrderNumber: "ORD-001",
		Status:      models.OrderStatusPending,
		Customer: &models.Customer{
			Name:  "John Doe",
			Phone: "+54 11 1234-5678",
			Email: "john@example.com",
			Address: &models.Address{
				Name:    "Av. Corrientes 1234",
				PlaceID: "ChIJxyz",
				Lat:     -34.6037,
				Lng:     -58.3816,
			},
		},
		PaymentMethod: &models.PaymentMethod{
			ID:   1,
			Name: "Cash",
			Code: models.PaymentMethodCash,
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID:   1,
			Name: "Delivery",
			Code: models.DeliveryMethodDelivery,
			DeliveryZones: []*models.DeliveryZone{
				{ID: 1, Name: "Zona Norte", Price: 300},
			},
		},
		Items: []*models.OrderItem{
			{
				Product: &models.Product{
					ID:               10,
					Name:             "Laptop",
					Price:            999.99,
					IsPromotional:    true,
					PromotionalPrice: 799.99,
					Images: []*models.Image{
						{URL: "https://example.com/laptop.jpg"},
					},
					Variants: []*models.Variant{
						{
							ID:   20,
							Name: "Color",
							Options: []*models.Option{
								{ID: 100, Name: "Silver", Price: 0},
							},
						},
						{
							ID:   21,
							Name: "RAM",
							Options: []*models.Option{
								{ID: 200, Name: "16GB", Price: 100},
							},
						},
					},
				},
				Quantity:   2,
				UnitPrice:  799.99,
				TotalPrice: 1599.98,
			},
		},
		ItemsCount:   2,
		Subtotal:     1599.98,
		ShippingCost: 300,
		Total:        1899.98,
		CreatedAt:    time.Date(2024, 5, 20, 12, 0, 0, 0, time.UTC),
	}
}

// =============================================================================
// OrderResponseFromModel Tests
// =============================================================================

func TestOrderResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := OrderResponseFromModel(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when full order then maps basic fields", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "ORD-001", result.OrderNumber)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, 2, result.ItemsCount)
		assert.Equal(t, 1599.98, result.Subtotal)
		assert.Equal(t, 300.0, result.ShippingCost)
		assert.Equal(t, 1899.98, result.Total)
		assert.Equal(t, time.Date(2024, 5, 20, 12, 0, 0, 0, time.UTC), result.CreatedAt)
	})

	t.Run("when order has customer with address then maps customer", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Equal(t, "John Doe", result.Customer.Name)
		assert.Equal(t, "+54 11 1234-5678", result.Customer.Phone)
		assert.Equal(t, "john@example.com", result.Customer.Email)
		assert.NotNil(t, result.Customer.Address)
		assert.Equal(t, "Av. Corrientes 1234", result.Customer.Address.Name)
		assert.Equal(t, "ChIJxyz", result.Customer.Address.PlaceID)
		assert.Equal(t, -34.6037, result.Customer.Address.Lat)
		assert.Equal(t, -58.3816, result.Customer.Address.Lng)
	})

	t.Run("when customer has no address then address is nil", func(t *testing.T) {
		// Arrange
		order := &models.Order{
			ID:       2,
			Customer: &models.Customer{Name: "Jane"},
			Items:    []*models.OrderItem{},
		}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Equal(t, "Jane", result.Customer.Name)
		assert.Nil(t, result.Customer.Address)
	})

	t.Run("when order has payment method then maps payment", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.NotNil(t, result.PaymentMethod)
		assert.Equal(t, 1, result.PaymentMethod.ID)
		assert.Equal(t, "Cash", result.PaymentMethod.Name)
		assert.Equal(t, "cash", result.PaymentMethod.Code)
	})

	t.Run("when no payment method then payment is nil", func(t *testing.T) {
		// Arrange
		order := &models.Order{ID: 3, Items: []*models.OrderItem{}}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Nil(t, result.PaymentMethod)
	})

	t.Run("when order has delivery method with zone then maps delivery", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.NotNil(t, result.DeliveryMethod)
		assert.Equal(t, 1, result.DeliveryMethod.ID)
		assert.Equal(t, "Delivery", result.DeliveryMethod.Name)
		assert.Equal(t, "delivery", result.DeliveryMethod.Code)
		assert.NotNil(t, result.DeliveryMethod.DeliveryZone)
		assert.Equal(t, 1, result.DeliveryMethod.DeliveryZone.ID)
		assert.Equal(t, "Zona Norte", result.DeliveryMethod.DeliveryZone.Name)
		assert.Equal(t, 300.0, result.DeliveryMethod.DeliveryZone.Price)
	})

	t.Run("when delivery method has no zones then zone is nil", func(t *testing.T) {
		// Arrange
		order := &models.Order{
			ID: 4,
			DeliveryMethod: &models.DeliveryMethod{
				ID:   2,
				Name: "Pickup",
				Code: models.DeliveryMethodPickup,
			},
			Items: []*models.OrderItem{},
		}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.NotNil(t, result.DeliveryMethod)
		assert.Nil(t, result.DeliveryMethod.DeliveryZone)
	})

	t.Run("when no delivery method then delivery is nil", func(t *testing.T) {
		// Arrange
		order := &models.Order{ID: 5, Items: []*models.OrderItem{}}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Nil(t, result.DeliveryMethod)
	})

	t.Run("when item has product with image then maps image_url", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "https://example.com/laptop.jpg", result.Items[0].Product.ImageURL)
	})

	t.Run("when item product has no images then image_url is empty", func(t *testing.T) {
		// Arrange
		order := &models.Order{
			ID: 6,
			Items: []*models.OrderItem{
				{
					Product:   &models.Product{ID: 1, Name: "No Image"},
					Quantity:  1,
					UnitPrice: 100,
				},
			},
		}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Equal(t, "", result.Items[0].Product.ImageURL)
	})

	t.Run("when item has selected options then flattens variants to selected_options", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Len(t, result.Items[0].SelectedOptions, 2)

		// Find options by variant name (map iteration order is not guaranteed)
		optionsByVariant := make(map[string]*OrderSelectedOptionResponse)
		for _, opt := range result.Items[0].SelectedOptions {
			optionsByVariant[opt.VariantName] = opt
		}

		colorOpt := optionsByVariant["Color"]
		assert.NotNil(t, colorOpt)
		assert.Equal(t, 20, colorOpt.VariantID)
		assert.Equal(t, "Color", colorOpt.VariantName)
		assert.Equal(t, 100, colorOpt.OptionID)
		assert.Equal(t, "Silver", colorOpt.OptionName)
		assert.Equal(t, 0.0, colorOpt.OptionPrice)

		ramOpt := optionsByVariant["RAM"]
		assert.NotNil(t, ramOpt)
		assert.Equal(t, 21, ramOpt.VariantID)
		assert.Equal(t, "RAM", ramOpt.VariantName)
		assert.Equal(t, 200, ramOpt.OptionID)
		assert.Equal(t, "16GB", ramOpt.OptionName)
		assert.Equal(t, 100.0, ramOpt.OptionPrice)
	})

	t.Run("when item has no variants then selected_options is empty", func(t *testing.T) {
		// Arrange
		order := &models.Order{
			ID: 7,
			Items: []*models.OrderItem{
				{
					Product:   &models.Product{ID: 1, Name: "Simple"},
					Quantity:  1,
					UnitPrice: 50,
				},
			},
		}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Len(t, result.Items[0].SelectedOptions, 0)
	})

	t.Run("when item maps product snapshot fields", func(t *testing.T) {
		// Arrange
		order := newTestOrder()

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		item := result.Items[0]
		assert.Equal(t, 10, item.Product.ID)
		assert.Equal(t, "Laptop", item.Product.Name)
		assert.Equal(t, 999.99, item.Product.Price)
		assert.True(t, item.Product.IsPromotional)
		assert.Equal(t, 799.99, item.Product.PromotionalPrice)
		assert.Equal(t, 2, item.Quantity)
		assert.Equal(t, 799.99, item.UnitPrice)
		assert.Equal(t, 1599.98, item.TotalPrice)
	})

	t.Run("when nil customer then customer fields are zero", func(t *testing.T) {
		// Arrange
		order := &models.Order{ID: 8, Items: []*models.OrderItem{}}

		// Act
		result := OrderResponseFromModel(order)

		// Assert
		assert.Equal(t, "", result.Customer.Name)
		assert.Nil(t, result.Customer.Address)
	})
}

// =============================================================================
// OrderResponsesFromModels Tests
// =============================================================================

func TestOrderResponsesFromModels(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		// Act
		result := OrderResponsesFromModels(nil)

		// Assert
		assert.Nil(t, result)
	})

	t.Run("when empty slice then returns empty slice", func(t *testing.T) {
		// Act
		result := OrderResponsesFromModels([]*models.Order{})

		// Assert
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("when multiple orders then maps all", func(t *testing.T) {
		// Arrange
		orders := []*models.Order{
			{ID: 1, OrderNumber: "ORD-001", Status: models.OrderStatusPending, Items: []*models.OrderItem{}},
			{ID: 2, OrderNumber: "ORD-002", Status: models.OrderStatusCompleted, Items: []*models.OrderItem{}},
		}

		// Act
		result := OrderResponsesFromModels(orders)

		// Assert
		assert.Len(t, result, 2)
		assert.Equal(t, "ORD-001", result[0].OrderNumber)
		assert.Equal(t, "ORD-002", result[1].OrderNumber)
	})
}

// =============================================================================
// NewListOrdersResponse Tests
// =============================================================================

func TestNewListOrdersResponse(t *testing.T) {
	t.Run("when orders with pagination then maps all fields", func(t *testing.T) {
		// Arrange
		orders := []*models.Order{
			{ID: 1, OrderNumber: "ORD-001", Items: []*models.OrderItem{}},
		}
		totalCount := 25

		// Act
		result := NewListOrdersResponse(orders, "cursor456", true, &totalCount)

		// Assert
		assert.Len(t, result.Orders, 1)
		assert.Equal(t, "cursor456", result.NextCursor)
		assert.True(t, result.HasMore)
		assert.Equal(t, 25, *result.TotalCount)
	})

	t.Run("when nil orders then orders is nil", func(t *testing.T) {
		// Act
		result := NewListOrdersResponse(nil, "", false, nil)

		// Assert
		assert.Nil(t, result.Orders)
		assert.False(t, result.HasMore)
	})
}
