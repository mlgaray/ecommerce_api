package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// GetBySlug Tests
// =============================================================================

func TestStoreService_GetBySlug(t *testing.T) {
	t.Run("when store exists then returns store", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-shop"
		expectedShop := &models.Shop{
			ID:    1,
			Name:  "Test Shop",
			Slug:  slug,
			Email: "test@shop.com",
			Images: []*models.Image{
				{ID: 1, URL: "https://cloudinary.com/logo.jpg", Type: "logo"},
			},
			Address: &models.Address{
				ID:   1,
				Name: "Main Street 123",
			},
			PaymentMethods: []*models.PaymentMethod{
				{ID: 1, Name: "Cash", Code: "cash", IsActive: true},
			},
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
			OperatingSchedules: []*models.OperatingSchedule{
				{ID: 1, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedShop.ID, result.ID)
		assert.Equal(t, expectedShop.Name, result.Name)
		assert.Equal(t, expectedShop.Slug, result.Slug)
		assert.Equal(t, expectedShop.Email, result.Email)
		assert.NotNil(t, result.Images)
		assert.NotNil(t, result.Address)
		assert.NotNil(t, result.PaymentMethods)
		assert.NotNil(t, result.DeliveryMethods)
		assert.NotNil(t, result.OperatingSchedules)
	})

	t.Run("when store not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "non-existent-shop"
		notFoundError := &errors.RecordNotFoundError{Message: errors.StoreNotFound}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
		assert.Equal(t, errors.StoreNotFound, notFound.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-shop"
		expectedError := stdErrors.New("database error")

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, expectedError)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when store has no images then returns store without images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-images"
		expectedShop := &models.Shop{
			ID:     1,
			Name:   "Shop Without Images",
			Slug:   slug,
			Images: []*models.Image{},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Images)
	})

	t.Run("when store has no address then returns store without address", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-address"
		expectedShop := &models.Shop{
			ID:      1,
			Name:    "Shop Without Address",
			Slug:    slug,
			Address: nil,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Address)
	})

	t.Run("when store has no payment methods then returns store without payment methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-payment-methods"
		expectedShop := &models.Shop{
			ID:             1,
			Name:           "Shop Without Payment Methods",
			Slug:           slug,
			PaymentMethods: []*models.PaymentMethod{},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.PaymentMethods)
	})

	t.Run("when store has no delivery methods then returns store without delivery methods", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-delivery-methods"
		expectedShop := &models.Shop{
			ID:              1,
			Name:            "Shop Without Delivery Methods",
			Slug:            slug,
			DeliveryMethods: []*models.DeliveryMethod{},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.DeliveryMethods)
	})

	t.Run("when store has no operating schedules then returns store without schedules", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "shop-without-schedules"
		expectedShop := &models.Shop{
			ID:                 1,
			Name:               "Shop Without Schedules",
			Slug:               slug,
			OperatingSchedules: []*models.OperatingSchedule{},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedShop, nil)

		productRepoMock := mocks.NewProductRepository(t)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		result, err := service.GetBySlug(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.OperatingSchedules)
	})
}

// =============================================================================
// CheckSlugAvailability Tests
// =============================================================================

func TestStoreService_CheckSlugAvailability(t *testing.T) {
	t.Run("when slug does not exist then returns true (available)", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "new-store"

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			SlugExists(ctx, slug).
			Return(false, nil)

		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		available, err := service.CheckSlugAvailability(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.True(t, available)
	})

	t.Run("when slug exists then returns false (not available)", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "existing-store"

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			SlugExists(ctx, slug).
			Return(true, nil)

		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		available, err := service.CheckSlugAvailability(ctx, slug)

		// Assert
		assert.NoError(t, err)
		assert.False(t, available)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "some-store"
		expectedError := stdErrors.New("database error")

		shopRepoMock := mocks.NewShopRepository(t)
		shopRepoMock.EXPECT().
			SlugExists(ctx, slug).
			Return(false, expectedError)

		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		available, err := service.CheckSlugAvailability(ctx, slug)

		// Assert
		assert.Error(t, err)
		assert.False(t, available)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// ValidateOrderItems Tests
// =============================================================================

func TestStoreService_ValidateOrderItems(t *testing.T) {
	t.Run("when item has no product then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{{Product: nil}}
		storeID := 1

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})

	t.Run("when product not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 999, Name: "Unknown", Price: 100}, Quantity: 1, UnitPrice: 100},
		}
		storeID := 1

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{999}, storeID).
			Return(map[int]*models.Product{}, nil) // Empty map = product not found

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var notFoundErr *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFoundErr))
	})

	t.Run("when product is not active then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Inactive Product", Price: 100}, Quantity: 1, UnitPrice: 100},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Inactive Product",
			Price:        100,
			IsActive:     false,
			IsStockeable: true,
			Stock:        10,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductNotActive, validationErr.Message)
	})

	t.Run("when product data mismatch then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Wrong Name", Price: 100}, Quantity: 1, UnitPrice: 100},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Correct Name",
			Price:        100,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductDataMismatch, validationErr.Message)
	})

	t.Run("when unit price mismatch without options then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Big Mac", Price: 10000}, Quantity: 1, UnitPrice: 15000}, // Wrong price
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.UnitPriceMismatch, validationErr.Message)
	})

	t.Run("when unit price mismatch with options then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		// Unit price should be 10000 + 1500 + 800 = 12300, but we send 15000
		items := []*models.OrderItem{
			{
				Product: &models.Product{
					ID:    1,
					Name:  "Big Mac",
					Price: 10000,
					Variants: []*models.Variant{
						{
							ID:   1,
							Name: "Tamaño",
							Options: []*models.Option{
								{ID: 2, Name: "Grande", Price: 1500},
							},
						},
						{
							ID:   2,
							Name: "Extras",
							Options: []*models.Option{
								{ID: 4, Name: "Extra queso", Price: 800},
							},
						},
					},
				},
				Quantity:  1,
				UnitPrice: 15000, // Wrong: should be 12300
			},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
			Variants: []*models.Variant{
				{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 1500}}},
				{ID: 2, Name: "Extras", Options: []*models.Option{{ID: 4, Name: "Extra queso", Price: 800}}},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.UnitPriceMismatch, validationErr.Message)
	})

	t.Run("when unit price matches with options then passes validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		// Unit price = 10000 + 1500 + 800 = 12300
		items := []*models.OrderItem{
			{
				Product: &models.Product{
					ID:    1,
					Name:  "Big Mac",
					Price: 10000,
					Variants: []*models.Variant{
						{
							ID:   1,
							Name: "Tamaño",
							Options: []*models.Option{
								{ID: 2, Name: "Grande", Price: 1500},
							},
						},
						{
							ID:   2,
							Name: "Extras",
							Options: []*models.Option{
								{ID: 4, Name: "Extra queso", Price: 800},
							},
						},
					},
				},
				Quantity:  1,
				UnitPrice: 12300, // Correct: 10000 + 1500 + 800
			},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
			Variants: []*models.Variant{
				{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 1500}}},
				{ID: 2, Name: "Extras", Options: []*models.Option{{ID: 4, Name: "Extra queso", Price: 800}}},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when promotional product unit price matches then passes validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		// Unit price = promotional 8000 + 1500 = 9500
		items := []*models.OrderItem{
			{
				Product: &models.Product{
					ID:               1,
					Name:             "Big Mac",
					Price:            10000,
					IsPromotional:    true,
					PromotionalPrice: 8000,
					Variants: []*models.Variant{
						{
							ID:   1,
							Name: "Tamaño",
							Options: []*models.Option{
								{ID: 2, Name: "Grande", Price: 1500},
							},
						},
					},
				},
				Quantity:  1,
				UnitPrice: 9500, // Correct: promotional 8000 + 1500
			},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    true,
			PromotionalPrice: 8000,
			IsActive:         true,
			IsStockeable:     true,
			Stock:            10,
			Variants: []*models.Variant{
				{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 1500}}},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when option has quantity greater than 1 then multiplies price by quantity", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		// Unit price = promotional 8500 + (1500*1) + (800*2) + (2000*1) = 8500 + 1500 + 1600 + 2000 = 13600
		items := []*models.OrderItem{
			{
				Product: &models.Product{
					ID:               1,
					Name:             "Big Mac",
					Price:            10000,
					IsPromotional:    true,
					PromotionalPrice: 8500,
					Variants: []*models.Variant{
						{
							ID:   1,
							Name: "Tamaño",
							Options: []*models.Option{
								{ID: 2, Name: "Grande", Price: 1500, Quantity: 1},
							},
						},
						{
							ID:   2,
							Name: "Ingredientes adicionales", //nolint:misspell // Spanish variant name
							Options: []*models.Option{
								{ID: 4, Name: "Extra queso", Price: 800, Quantity: 2},
							},
						},
						{
							ID:   3,
							Name: "Bebida",
							Options: []*models.Option{
								{ID: 8, Name: "Coca Cola", Price: 2000, Quantity: 1},
							},
						},
					},
				},
				Quantity:  2,
				UnitPrice: 13600, // Correct: 8500 + 1500 + 1600 + 2000
			},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:               1,
			Name:             "Big Mac",
			Price:            10000,
			IsPromotional:    true,
			PromotionalPrice: 8500,
			IsActive:         true,
			IsStockeable:     true,
			Stock:            10,
			Variants: []*models.Variant{
				{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 1500}}},
				{ID: 2, Name: "Ingredientes adicionales", Options: []*models.Option{{ID: 4, Name: "Extra queso", Price: 800}}}, //nolint:misspell // Spanish variant name
				{ID: 3, Name: "Bebida", Options: []*models.Option{{ID: 8, Name: "Coca Cola", Price: 2000}}},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when insufficient stock then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Big Mac", Price: 10000}, Quantity: 100, UnitPrice: 10000},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        5, // Only 5 in stock
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var businessErr *errors.BusinessRuleError
		assert.True(t, stdErrors.As(err, &businessErr))
		assert.Equal(t, errors.InsufficientStock, businessErr.Message)
	})

	t.Run("when all validations pass then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Big Mac", Price: 10000}, Quantity: 2, UnitPrice: 10000},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when item has product with ID zero then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{{Product: &models.Product{ID: 0}}}
		storeID := 1

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderItemProductRequired, validationErr.Message)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Product", Price: 100}, Quantity: 1, UnitPrice: 100},
		}
		storeID := 1
		expectedError := stdErrors.New("database error")

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(nil, expectedError)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when variant data mismatch then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{
				Product: &models.Product{
					ID:    1,
					Name:  "Big Mac",
					Price: 10000,
					Variants: []*models.Variant{
						{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 1500}}},
					},
				},
				Quantity:  1,
				UnitPrice: 11500,
			},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: true,
			Stock:        10,
			Variants: []*models.Variant{
				{ID: 1, Name: "Tamaño", Options: []*models.Option{{ID: 2, Name: "Grande", Price: 2000}}}, // Different price
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductDataMismatch, validationErr.Message)
	})

	t.Run("when non-stockeable product then stock validation passes", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Big Mac", Price: 10000}, Quantity: 100, UnitPrice: 10000},
		}
		storeID := 1

		dbProduct := &models.Product{
			ID:           1,
			Name:         "Big Mac",
			Price:        10000,
			IsActive:     true,
			IsStockeable: false, // Doesn't manage stock
			Stock:        0,
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1}, storeID).
			Return(map[int]*models.Product{1: dbProduct}, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when multiple items all valid then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		items := []*models.OrderItem{
			{Product: &models.Product{ID: 1, Name: "Big Mac", Price: 10000}, Quantity: 2, UnitPrice: 10000},
			{Product: &models.Product{ID: 2, Name: "McFlurry", Price: 5000}, Quantity: 1, UnitPrice: 5000},
		}
		storeID := 1

		productsMap := map[int]*models.Product{
			1: {ID: 1, Name: "Big Mac", Price: 10000, IsActive: true, IsStockeable: false},
			2: {ID: 2, Name: "McFlurry", Price: 5000, IsActive: true, IsStockeable: false},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		productRepoMock.EXPECT().
			GetByIDsAndShopID(ctx, []int{1, 2}, storeID).
			Return(productsMap, nil)

		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateOrderItems(ctx, items, storeID)

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// ValidateDeliveryMethod Tests
// =============================================================================

func TestStoreService_ValidateDeliveryMethod(t *testing.T) {
	t.Run("when delivery method is nil then returns validation error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, nil, 0)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderDeliveryMethodRequired, validationErr.Message)
	})

	t.Run("when delivery method ID is zero then returns validation error", func(t *testing.T) {
		// Arrange
		store := &models.Store{}
		deliveryMethod := &models.DeliveryMethod{ID: 0}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.OrderDeliveryMethodRequired, validationErr.Message)
	})

	t.Run("when delivery method not found in store then returns validation error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 999}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.DeliveryMethodNotFound, validationErr.Message)
	})

	t.Run("when delivery method is inactive then returns validation error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: false},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.DeliveryMethodNotFound, validationErr.Message)
	})

	t.Run("when pickup with zero cost then passes", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Pickup", Code: models.DeliveryMethodPickup, IsActive: true},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when pickup with non-zero cost then returns error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Pickup", Code: models.DeliveryMethodPickup, IsActive: true},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 500)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PickupShouldHaveZeroCost, validationErr.Message)
	})

	t.Run("when fixed price matches then passes", func(t *testing.T) {
		// Arrange
		fixedPrice := 500.0
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryConfig: &models.DeliveryConfig{FixedPrice: &fixedPrice},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 500)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when fixed price mismatch then returns error", func(t *testing.T) {
		// Arrange
		fixedPrice := 500.0
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryConfig: &models.DeliveryConfig{FixedPrice: &fixedPrice},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 999)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ShippingCostMismatch, validationErr.Message)
	})

	t.Run("when zone-based and price matches then passes", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryZones: []*models.DeliveryZone{
						{ID: 10, Name: "Zone A", Price: 300},
						{ID: 20, Name: "Zone B", Price: 500},
					},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{
			ID: 1,
			DeliveryZones: []*models.DeliveryZone{
				{ID: 20}, // Selected zone B
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 500)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when zone-based and no zone selected then returns error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryZones: []*models.DeliveryZone{
						{ID: 10, Name: "Zone A", Price: 300},
					},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1} // No zones selected

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 300)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.DeliveryZoneRequired, validationErr.Message)
	})

	t.Run("when zone-based and zone not found then returns error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryZones: []*models.DeliveryZone{
						{ID: 10, Name: "Zone A", Price: 300},
					},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{
			ID: 1,
			DeliveryZones: []*models.DeliveryZone{
				{ID: 999}, // Non-existent zone
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 300)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.DeliveryZoneNotFound, validationErr.Message)
	})

	t.Run("when zone-based and price mismatch then returns error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{
					ID: 1, Name: "Delivery", Code: "delivery", IsActive: true,
					DeliveryZones: []*models.DeliveryZone{
						{ID: 10, Name: "Zone A", Price: 300},
					},
				},
			},
		}
		deliveryMethod := &models.DeliveryMethod{
			ID: 1,
			DeliveryZones: []*models.DeliveryZone{
				{ID: 10},
			},
		}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 999)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ShippingCostMismatch, validationErr.Message)
	})

	t.Run("when no config and zero cost then passes (free delivery)", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when no config and non-zero cost then returns error", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{
				{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
			},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 500)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ShippingCostMismatch, validationErr.Message)
	})

	t.Run("when store has no delivery methods then returns not found", func(t *testing.T) {
		// Arrange
		store := &models.Store{
			DeliveryMethods: []*models.DeliveryMethod{},
		}
		deliveryMethod := &models.DeliveryMethod{ID: 1}

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		service := NewStoreService(shopRepoMock, productRepoMock, nil)

		// Act
		err := service.ValidateDeliveryMethod(store, deliveryMethod, 0)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.DeliveryMethodNotFound, validationErr.Message)
	})
}

// =============================================================================
// RecordVisit Tests
// =============================================================================

func TestStoreService_RecordVisit(t *testing.T) {
	t.Run("when repository succeeds then returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		storeVisitRepoMock := mocks.NewStoreVisitRepository(t)
		storeVisitRepoMock.EXPECT().RecordVisit(ctx, shopID).Return(nil)

		service := NewStoreService(shopRepoMock, productRepoMock, storeVisitRepoMock)

		// Act
		err := service.RecordVisit(ctx, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when repository errors then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		expectedError := stdErrors.New("database error")

		shopRepoMock := mocks.NewShopRepository(t)
		productRepoMock := mocks.NewProductRepository(t)
		storeVisitRepoMock := mocks.NewStoreVisitRepository(t)
		storeVisitRepoMock.EXPECT().RecordVisit(ctx, shopID).Return(expectedError)

		service := NewStoreService(shopRepoMock, productRepoMock, storeVisitRepoMock)

		// Act
		err := service.RecordVisit(ctx, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})
}
