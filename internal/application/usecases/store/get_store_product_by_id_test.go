package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidProduct creates a valid active product for testing
func newValidProduct() *models.Product {
	return &models.Product{
		ID:          1,
		Name:        "Test Product",
		Description: "Test product description",
		Price:       10.99,
		Stock:       100,
		IsActive:    true,
		Category: &models.Category{
			ID:   1,
			Name: "Test Category",
		},
		Images: []*models.Image{
			{ID: 1, URL: "https://cloudinary.com/product.jpg"},
		},
		Variants: []*models.Variant{
			{
				ID:            1,
				Name:          "Size",
				Order:         1,
				SelectionType: "single",
				MaxSelections: 1,
				IsRequired:    true,
				Options: []*models.Option{
					{ID: 1, Name: "Small", Price: 0, Order: 1},
					{ID: 2, Name: "Large", Price: 2.00, Order: 2},
				},
			},
		},
	}
}

// newInactiveProduct creates an inactive product for testing
func newInactiveProduct() *models.Product {
	product := newValidProduct()
	product.IsActive = false
	return product
}

// =============================================================================
// GetStoreProductByIDUseCase Tests
// =============================================================================

func TestGetStoreProductByIDUseCase_Execute(t *testing.T) {
	t.Run("when store and product exist and product is active then returns product", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1
		expectedStore := newValidStore()
		expectedProduct := newValidProduct()

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(expectedProduct, nil)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedProduct.ID, result.ID)
		assert.Equal(t, expectedProduct.Name, result.Name)
		assert.True(t, result.IsActive)
		assert.NotNil(t, result.Category)
		assert.NotNil(t, result.Images)
		assert.NotNil(t, result.Variants)
		assert.Len(t, result.Variants, 1)
		assert.Len(t, result.Variants[0].Options, 2)
	})

	t.Run("when store not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "non-existent-store"
		productID := 1
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, notFoundError)

		productServiceMock := mocks.NewProductService(t)
		// ProductService should not be called when store is not found

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.StoreNotFound, notFound.Message)
	})

	t.Run("when product not found then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 99999
		expectedStore := newValidStore()
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ProductNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(nil, notFoundError)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ProductNotFound, notFound.Message)
	})

	t.Run("when product belongs to another shop then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1 // Product exists but belongs to shop 2, not shop 1
		expectedStore := newValidStore()
		notFoundError := &domainErrors.RecordNotFoundError{Message: domainErrors.ProductNotFound}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(nil, notFoundError) // Repository returns not found when shop_id doesn't match

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ProductNotFound, notFound.Message)
	})

	t.Run("when product is inactive then returns RecordNotFoundError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1
		expectedStore := newValidStore()
		inactiveProduct := newInactiveProduct()

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(inactiveProduct, nil)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFound *domainErrors.RecordNotFoundError
		assert.True(t, errors.As(err, &notFound))
		assert.Equal(t, domainErrors.ProductNotFound, notFound.Message)
	})

	t.Run("when store service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1
		expectedError := errors.New("database error")

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(nil, expectedError)

		productServiceMock := mocks.NewProductService(t)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when product service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1
		expectedStore := newValidStore()
		expectedError := errors.New("database error")

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(nil, expectedError)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when product has no variants then returns product without variants", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		slug := "test-store"
		productID := 1
		expectedStore := newValidStore()
		productWithoutVariants := &models.Product{
			ID:       1,
			Name:     "Simple Product",
			Price:    5.99,
			IsActive: true,
			Category: &models.Category{ID: 1, Name: "Test"},
			Variants: []*models.Variant{},
		}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, slug).
			Return(expectedStore, nil)

		productServiceMock := mocks.NewProductService(t)
		productServiceMock.EXPECT().
			GetByIDAndShopID(ctx, productID, expectedStore.ID).
			Return(productWithoutVariants, nil)

		useCase := NewGetStoreProductByIDUseCase(storeServiceMock, productServiceMock)

		// Act
		result, err := useCase.Execute(ctx, slug, productID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Variants)
	})
}
