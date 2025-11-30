package services

import (
	"context"
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newValidProduct creates a valid product for testing
func newValidProduct() *models.Product {
	return &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       100.00,
		Stock:       10,
		IsActive:    true,
	}
}

// newValidProductWithID creates a valid product with ID for testing
func newValidProductWithID(id int) *models.Product {
	p := newValidProduct()
	p.ID = id
	return p
}

// newValidFilters creates valid filters for testing with default shop ID
func newValidFilters() models.ProductFilters {
	return models.ProductFilters{
		ShopID: 1,
		Limit:  20,
	}
}

// =============================================================================
// Create Tests
// =============================================================================

func TestProductService_Create(t *testing.T) {
	t.Run("when product is valid then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		imageBuffers := [][]byte{[]byte("image1"), []byte("image2")}

		expectedProduct := newValidProductWithID(1)
		expectedProduct.Images = []*models.Image{
			{URL: "https://cloudinary.com/image1.jpg", StorageRef: "shop_1/products/abc1"},
			{URL: "https://cloudinary.com/image2.jpg", StorageRef: "shop_1/products/abc2"},
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			UploadMultiple(ctx, imageBuffers, "shop_1/products").
			Return([]*models.Image{
				{URL: "https://cloudinary.com/image1.jpg", StorageRef: "shop_1/products/abc1"},
				{URL: "https://cloudinary.com/image2.jpg", StorageRef: "shop_1/products/abc2"},
			}, nil)

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(expectedProduct, nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, imageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedProduct.ID, result.ID)
	})

	t.Run("when product is valid without images then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		var imageBuffers [][]byte // No images

		expectedProduct := newValidProductWithID(1)

		assetMock := mocks.NewAssetService(t)
		// AssetService should NOT be called when no images

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(expectedProduct, nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, imageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, product.Images)
	})

	t.Run("when price is zero then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Price = 0 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("when price is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Price = -50.00 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("when stock is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Stock = -5 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductStockCannotBeNegative, validationErr.Message)
	})

	t.Run("when minimum stock is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.MinimumStock = -1 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductMinimumStockCannotBeNegative, validationErr.Message)
	})

	t.Run("when minimum stock greater than stock then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Stock = 5
		product.MinimumStock = 10 // Invalid: greater than stock

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductMinimumStockCannotBeGreaterThanStock, validationErr.Message)
	})

	t.Run("when minimum stock set but stock is zero then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Stock = 0
		product.MinimumStock = 5 // Invalid: requires stock

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.MinimumStockRequiresStock, validationErr.Message)
	})

	t.Run("when promotional without promotional price then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.IsPromotional = true
		product.PromotionalPrice = 0 // Invalid: promotional requires price

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalProductRequiresPromotionalPrice, validationErr.Message)
	})

	t.Run("when promotional price greater than regular price then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 150.00 // Invalid: greater than regular

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalPriceMustBeLowerThanRegularPrice, validationErr.Message)
	})

	t.Run("when promotional price equals regular price then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 100.00 // Invalid: equal to regular

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.PromotionalPriceMustBeLowerThanRegularPrice, validationErr.Message)
	})

	t.Run("when valid promotional product then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Price = 100.00
		product.IsPromotional = true
		product.PromotionalPrice = 75.00 // Valid: lower than regular

		expectedProduct := newValidProductWithID(1)
		expectedProduct.IsPromotional = true
		expectedProduct.PromotionalPrice = 75.00

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(expectedProduct, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		expectedError := stdErrors.New("database connection failed")

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when images provided then uploads via AssetService", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		imageBuffers := [][]byte{[]byte("img1"), []byte("img2"), []byte("img3")}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			UploadMultiple(ctx, imageBuffers, "shop_1/products").
			Return([]*models.Image{
				{URL: "https://cloudinary.com/img1.jpg", StorageRef: "shop_1/products/abc1"},
				{URL: "https://cloudinary.com/img2.jpg", StorageRef: "shop_1/products/abc2"},
				{URL: "https://cloudinary.com/img3.jpg", StorageRef: "shop_1/products/abc3"},
			}, nil)

		var capturedProduct *models.Product
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Run(func(ctx context.Context, p *models.Product, s int) {
				capturedProduct = p
			}).
			Return(newValidProductWithID(1), nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.Create(ctx, product, imageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, capturedProduct.Images, 3)
		assert.Equal(t, "https://cloudinary.com/img1.jpg", capturedProduct.Images[0].URL)
		assert.Equal(t, "https://cloudinary.com/img2.jpg", capturedProduct.Images[1].URL)
		assert.Equal(t, "https://cloudinary.com/img3.jpg", capturedProduct.Images[2].URL)
	})
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestProductService_GetByID(t *testing.T) {
	t.Run("when product exists then returns product", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		expectedProduct := newValidProductWithID(productID)

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, productID).
			Return(expectedProduct, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, productID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, productID, result.ID)
	})

	t.Run("when product not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 999
		expectedError := &errors.RecordNotFoundError{Message: errors.ProductNotFound}

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, productID).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var notFoundErr *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFoundErr))
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		expectedError := stdErrors.New("database timeout")

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, productID).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetByID(ctx, productID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// Update Tests
// =============================================================================

func TestProductService_Update(t *testing.T) {
	t.Run("when product is valid then updates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Update(ctx, productID, mock.AnythingOfType("*models.Product")).
			Return([]string{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, nil, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when product is valid with new images then updates with appended images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.Images = []*models.Image{
			{ID: 1, URL: "https://existing.com/image1"},
		}
		newImageBuffers := [][]byte{[]byte("new_img1"), []byte("new_img2")}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			UploadMultiple(ctx, newImageBuffers, "shop_1/products").
			Return([]*models.Image{
				{URL: "https://cloudinary.com/new1.jpg", StorageRef: "shop_1/products/new1"},
				{URL: "https://cloudinary.com/new2.jpg", StorageRef: "shop_1/products/new2"},
			}, nil)

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Update(ctx, productID, mock.AnythingOfType("*models.Product")).
			Return([]string{}, nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, newImageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, product.Images, 3) // 1 existing + 2 new
	})

	t.Run("when price is invalid then returns validation error without calling repository", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.Price = -10.00 // Invalid

		repoMock := mocks.NewProductRepository(t)
		// Repository should NOT be called
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductPriceMustBePositive, validationErr.Message)
	})

	t.Run("when stock validation fails then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.Stock = -1 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ProductStockCannotBeNegative, validationErr.Message)
	})

	t.Run("when promotional validation fails then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.IsPromotional = true
		product.PromotionalPrice = 0 // Invalid

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		expectedError := stdErrors.New("constraint violation")

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Update(ctx, productID, mock.AnythingOfType("*models.Product")).
			Return([]string{}, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, nil, shopID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when empty image buffers then does not append images", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.Images = []*models.Image{
			{ID: 1, URL: "https://existing.com/image1"},
		}
		var emptyBuffers [][]byte

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Update(ctx, productID, mock.AnythingOfType("*models.Product")).
			Return([]string{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, emptyBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, product.Images, 1) // Only existing image
	})
}

// =============================================================================
// GetAllByShopIDWithFilters Tests
// =============================================================================

func TestProductService_GetAllByShopIDWithFilters(t *testing.T) {
	t.Run("when filters are valid then returns products", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()
		expectedProducts := []*models.Product{
			newValidProductWithID(1),
			newValidProductWithID(2),
		}

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return(expectedProducts, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("when shop_id is zero then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: 0, // Invalid
			Limit:  20,
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.ShopIDIsRequired, validationErr.Message)
	})

	t.Run("when shop_id is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: -1, // Invalid
			Limit:  20,
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
	})

	t.Run("when min_price is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		minPrice := -10.0
		filters := models.ProductFilters{
			ShopID:   1,
			Limit:    20,
			MinPrice: &minPrice, // Invalid
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.MinPriceCannotBeNegative, validationErr.Message)
	})

	t.Run("when max_price is negative then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		maxPrice := -10.0
		filters := models.ProductFilters{
			ShopID:   1,
			Limit:    20,
			MaxPrice: &maxPrice, // Invalid
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.MaxPriceCannotBeNegative, validationErr.Message)
	})

	t.Run("when min_price greater than max_price then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		minPrice := 100.0
		maxPrice := 50.0
		filters := models.ProductFilters{
			ShopID:   1,
			Limit:    20,
			MinPrice: &minPrice,
			MaxPrice: &maxPrice, // Invalid: max < min
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.MinPriceCannotBeGreaterThanMaxPrice, validationErr.Message)
	})

	t.Run("when invalid sort field then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: 1,
			Limit:  20,
			SortBy: "invalid_field", // Invalid
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
	})

	t.Run("when invalid sort order then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID:    1,
			Limit:     20,
			SortOrder: "invalid", // Invalid
		}

		repoMock := mocks.NewProductRepository(t)
		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &validationErr))
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
	})

	t.Run("when valid sort fields then passes validation", func(t *testing.T) {
		validSortFields := []string{"price", "name", "created_at", ""}
		for _, sortBy := range validSortFields {
			t.Run("sort_by_"+sortBy, func(t *testing.T) {
				// Arrange
				ctx := context.Background()
				filters := models.ProductFilters{
					ShopID: 1,
					Limit:  20,
					SortBy: sortBy,
				}

				repoMock := mocks.NewProductRepository(t)
				repoMock.EXPECT().
					GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
					Return([]*models.Product{}, nil)

				assetMock := mocks.NewAssetService(t)
				service := NewProductService(repoMock, assetMock)

				// Act
				_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

				// Assert
				assert.NoError(t, err)
			})
		}
	})

	t.Run("when valid sort orders then passes validation", func(t *testing.T) {
		validSortOrders := []string{"asc", "desc", ""}
		for _, sortOrder := range validSortOrders {
			t.Run("sort_order_"+sortOrder, func(t *testing.T) {
				// Arrange
				ctx := context.Background()
				filters := models.ProductFilters{
					ShopID:    1,
					Limit:     20,
					SortOrder: sortOrder,
				}

				repoMock := mocks.NewProductRepository(t)
				repoMock.EXPECT().
					GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
					Return([]*models.Product{}, nil)

				assetMock := mocks.NewAssetService(t)
				service := NewProductService(repoMock, assetMock)

				// Act
				_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

				// Assert
				assert.NoError(t, err)
			})
		}
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()
		expectedError := stdErrors.New("database error")

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return(nil, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})

	t.Run("when no products found then returns empty slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("when limit is zero then defaults to 20", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: 1,
			Limit:  0, // Should default to 20
		}

		var capturedFilters models.ProductFilters
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Run(func(ctx context.Context, f models.ProductFilters) {
				capturedFilters = f
			}).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 20, capturedFilters.Limit)
	})

	t.Run("when limit exceeds 100 then caps at 100", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: 1,
			Limit:  500, // Should cap at 100
		}

		var capturedFilters models.ProductFilters
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Run(func(ctx context.Context, f models.ProductFilters) {
				capturedFilters = f
			}).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 100, capturedFilters.Limit)
	})

	t.Run("when sort_by is empty then defaults to created_at", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID: 1,
			Limit:  20,
			SortBy: "", // Should default to created_at
		}

		var capturedFilters models.ProductFilters
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Run(func(ctx context.Context, f models.ProductFilters) {
				capturedFilters = f
			}).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "created_at", capturedFilters.SortBy)
	})

	t.Run("when sort_order is empty then defaults to desc", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID:    1,
			Limit:     20,
			SortOrder: "", // Should default to desc
		}

		var capturedFilters models.ProductFilters
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Run(func(ctx context.Context, f models.ProductFilters) {
				capturedFilters = f
			}).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "desc", capturedFilters.SortOrder)
	})
}

// =============================================================================
// CountByShopIDWithFilters Tests
// =============================================================================

func TestProductService_CountByShopIDWithFilters(t *testing.T) {
	t.Run("when filters are valid then returns count", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()
		expectedCount := 42

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, filters).
			Return(expectedCount, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCount, result)
	})

	t.Run("when no products then returns zero", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, filters).
			Return(0, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, filters)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("when repository returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := newValidFilters()
		expectedError := stdErrors.New("count query failed")

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			CountByShopIDWithFilters(ctx, filters).
			Return(0, expectedError)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.CountByShopIDWithFilters(ctx, filters)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// Private Method Tests (via public interface)
// =============================================================================

func TestProductService_prepareImagesForCreate(t *testing.T) {
	t.Run("when multiple images then uploads via AssetService", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		imageBuffers := [][]byte{
			[]byte("image_data_1"),
			[]byte("image_data_2"),
			[]byte("image_data_3"),
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			UploadMultiple(ctx, imageBuffers, "shop_1/products").
			Return([]*models.Image{
				{URL: "https://cloudinary.com/image1.jpg", StorageRef: "shop_1/products/ref1"},
				{URL: "https://cloudinary.com/image2.jpg", StorageRef: "shop_1/products/ref2"},
				{URL: "https://cloudinary.com/image3.jpg", StorageRef: "shop_1/products/ref3"},
			}, nil)

		var capturedProduct *models.Product
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Run(func(ctx context.Context, p *models.Product, s int) {
				capturedProduct = p
			}).
			Return(newValidProductWithID(1), nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.Create(ctx, product, imageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, capturedProduct.Images, 3)
		// Verify each image has URL and StorageRef from AssetService
		for i, img := range capturedProduct.Images {
			assert.Contains(t, img.URL, "https://cloudinary.com/image")
			assert.NotEmpty(t, img.StorageRef)
			assert.Equal(t, 0, img.ID) // New images have no ID yet
			_ = i
		}
	})

	t.Run("when no images then product has empty images slice", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()

		var capturedProduct *models.Product
		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Run(func(ctx context.Context, p *models.Product, s int) {
				capturedProduct = p
			}).
			Return(newValidProductWithID(1), nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, capturedProduct.Images)
	})
}

func TestProductService_prepareImagesForUpdate(t *testing.T) {
	t.Run("when new images added then appends to existing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		productID := 1
		shopID := 1
		product := newValidProduct()
		product.Images = []*models.Image{
			{ID: 100, URL: "https://cdn.example.com/existing1.jpg"},
			{ID: 101, URL: "https://cdn.example.com/existing2.jpg"},
		}
		newImageBuffers := [][]byte{
			[]byte("new_image_data"),
		}

		assetMock := mocks.NewAssetService(t)
		assetMock.EXPECT().
			UploadMultiple(ctx, newImageBuffers, "shop_1/products").
			Return([]*models.Image{
				{URL: "https://cloudinary.com/new.jpg", StorageRef: "shop_1/products/new123"},
			}, nil)

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Update(ctx, productID, mock.AnythingOfType("*models.Product")).
			Return([]string{}, nil)

		service := NewProductService(repoMock, assetMock)

		// Act
		err := service.Update(ctx, productID, product, newImageBuffers, shopID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, product.Images, 3)
		// First two are existing images
		assert.Equal(t, 100, product.Images[0].ID)
		assert.Equal(t, 101, product.Images[1].ID)
		// Third is new image
		assert.Equal(t, 0, product.Images[2].ID)
		assert.Equal(t, "https://cloudinary.com/new.jpg", product.Images[2].URL)
	})
}

// =============================================================================
// Edge Cases & Regression Tests
// =============================================================================

func TestProductService_EdgeCases(t *testing.T) {
	t.Run("when product has zero stock and zero minimum_stock then is valid", func(t *testing.T) {
		// Arrange - This is a valid case (out of stock product)
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.Stock = 0
		product.MinimumStock = 0

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(newValidProductWithID(1), nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when non-promotional product has promotional_price then ignores it", func(t *testing.T) {
		// Arrange - promotional_price should be ignored when IsPromotional is false
		ctx := context.Background()
		shopID := 1
		product := newValidProduct()
		product.IsPromotional = false
		product.PromotionalPrice = 50.00 // Should be ignored

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			Create(ctx, mock.AnythingOfType("*models.Product"), shopID).
			Return(newValidProductWithID(1), nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		result, err := service.Create(ctx, product, nil, shopID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when filters have all optional fields nil then is valid", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.ProductFilters{
			ShopID:        1,
			Limit:         20,
			Search:        nil,
			CategoryID:    nil,
			IsActive:      nil,
			IsHighlighted: nil,
			IsPromotional: nil,
			MinPrice:      nil,
			MaxPrice:      nil,
			LastID:        nil,
			LastSortValue: nil,
		}

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when filters have all optional fields set then is valid", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		search := "test"
		categoryID := 5
		isActive := true
		isHighlighted := false
		isPromotional := true
		minPrice := 10.0
		maxPrice := 100.0
		lastID := 50

		filters := models.ProductFilters{
			ShopID:        1,
			Limit:         50,
			Search:        &search,
			CategoryID:    &categoryID,
			IsActive:      &isActive,
			IsHighlighted: &isHighlighted,
			IsPromotional: &isPromotional,
			MinPrice:      &minPrice,
			MaxPrice:      &maxPrice,
			SortBy:        "price",
			SortOrder:     "asc",
			LastID:        &lastID,
		}

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when price range with equal values then is valid", func(t *testing.T) {
		// Arrange - min_price == max_price is valid (exact price match)
		ctx := context.Background()
		price := 50.0
		filters := models.ProductFilters{
			ShopID:   1,
			Limit:    20,
			MinPrice: &price,
			MaxPrice: &price,
		}

		repoMock := mocks.NewProductRepository(t)
		repoMock.EXPECT().
			GetAllByShopIDWithFilters(ctx, mock.AnythingOfType("models.ProductFilters")).
			Return([]*models.Product{}, nil)

		assetMock := mocks.NewAssetService(t)
		service := NewProductService(repoMock, assetMock)

		// Act
		_, err := service.GetAllByShopIDWithFilters(ctx, &filters)

		// Assert
		assert.NoError(t, err)
	})
}
