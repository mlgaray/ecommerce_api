package coupon

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestCoupon() *models.Coupon {
	return &models.Coupon{
		Code:           "SAVE10",
		Type:           models.CouponTypePercentage,
		Value:          10,
		MinOrderAmount: 50,
		IsActive:       true,
	}
}

func newTestCouponWithID(id int) *models.Coupon {
	c := newTestCoupon()
	c.ID = id
	c.ShopID = 1
	c.UsageCount = 3
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return c
}

func newTestCoupons() []*models.Coupon {
	return []*models.Coupon{
		newTestCouponWithID(1),
		newTestCouponWithID(2),
	}
}

// mockPaginationService is a local mock for ports.PaginationService[*models.Coupon]
// since the generic mock requires type instantiation.
type mockPaginationService struct {
	mock.Mock
}

func (m *mockPaginationService) ApplyPagination(items []*models.Coupon, limit int, sortBy string) ([]*models.Coupon, string, bool) {
	args := m.Called(items, limit, sortBy)
	coupons, _ := args.Get(0).([]*models.Coupon)
	return coupons, args.String(1), args.Bool(2)
}

func newMockPaginationService(t *testing.T) *mockPaginationService {
	m := &mockPaginationService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

// =============================================================================
// CreateCouponUseCase Tests
// =============================================================================

func TestCreateCouponUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns created coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		input := newTestCoupon()
		expected := newTestCouponWithID(1)

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Create(ctx, input, shopID).
			Return(expected, nil)

		useCase := NewCreateCouponUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, input, shopID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		input := newTestCoupon()
		expectedError := &errors.ValidationError{Message: errors.CouponCodeIsRequired}

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Create(ctx, input, shopID).
			Return(nil, expectedError)

		useCase := NewCreateCouponUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, input, shopID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})
}

// =============================================================================
// UpdateCouponUseCase Tests
// =============================================================================

func TestUpdateCouponUseCase_Execute(t *testing.T) {
	t.Run("when update and get succeed then returns updated coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 1
		shopID := 1
		input := newTestCoupon()
		expected := newTestCouponWithID(couponID)

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Update(ctx, couponID, shopID, input).
			Return(nil)
		couponServiceMock.EXPECT().
			GetByID(ctx, couponID, shopID).
			Return(expected, nil)

		useCase := NewUpdateCouponUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, couponID, shopID, input)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when update fails then returns error without calling GetByID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 1
		shopID := 1
		input := newTestCoupon()
		expectedError := &errors.RecordNotFoundError{Message: "coupon not found"}

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Update(ctx, couponID, shopID, input).
			Return(expectedError)
		// GetByID should NOT be called - testify will fail if it is

		useCase := NewUpdateCouponUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, couponID, shopID, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when update succeeds but GetByID fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 1
		shopID := 1
		input := newTestCoupon()
		expectedError := stdErrors.New("database error on re-fetch")

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Update(ctx, couponID, shopID, input).
			Return(nil)
		couponServiceMock.EXPECT().
			GetByID(ctx, couponID, shopID).
			Return(nil, expectedError)

		useCase := NewUpdateCouponUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, couponID, shopID, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
	})
}

// =============================================================================
// DeleteCouponUseCase Tests
// =============================================================================

func TestDeleteCouponUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns nil error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 1
		shopID := 1

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Delete(ctx, couponID, shopID).
			Return(nil)

		useCase := NewDeleteCouponUseCase(couponServiceMock)

		// Act
		err := useCase.Execute(ctx, couponID, shopID)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 99
		shopID := 1
		expectedError := &errors.RecordNotFoundError{Message: "coupon not found"}

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			Delete(ctx, couponID, shopID).
			Return(expectedError)

		useCase := NewDeleteCouponUseCase(couponServiceMock)

		// Act
		err := useCase.Execute(ctx, couponID, shopID)

		// Assert
		require.Error(t, err)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// GetCouponByIDUseCase Tests
// =============================================================================

func TestGetCouponByIDUseCase_Execute(t *testing.T) {
	t.Run("when service succeeds then returns coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 1
		shopID := 1
		expected := newTestCouponWithID(couponID)

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			GetByID(ctx, couponID, shopID).
			Return(expected, nil)

		useCase := NewGetCouponByIDUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, couponID, shopID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when service returns error then propagates error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		couponID := 99
		shopID := 1
		expectedError := &errors.RecordNotFoundError{Message: "coupon not found"}

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			GetByID(ctx, couponID, shopID).
			Return(nil, expectedError)

		useCase := NewGetCouponByIDUseCase(couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, couponID, shopID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})
}

// =============================================================================
// GetAllCouponsByShopIDUseCase Tests
// =============================================================================

func TestGetAllCouponsByShopIDUseCase_Execute_Validation(t *testing.T) {
	t.Run("when invalid sort field then returns ValidationError", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CouponFilters{
			SortBy: "invalid_field",
		}

		couponServiceMock := mocks.NewCouponService(t)
		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
		assert.Equal(t, errors.InvalidSortField, valErr.Message)
		assert.Nil(t, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllCouponsByShopIDUseCase_Execute_FirstPage(t *testing.T) {
	shopID := 1

	t.Run("when first page succeeds then returns coupons with totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.CouponFilters{} // LastID is nil = first page
		expectedCoupons := newTestCoupons()
		expectedCount := 10

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCount, nil)
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCoupons, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", expectedCoupons, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedCoupons, "", false)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedCoupons, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})

	t.Run("when count fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.CouponFilters{}
		expectedError := stdErrors.New("count query failed")

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(0, expectedError)
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, nil)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when select fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		filters := models.CouponFilters{}
		expectedError := stdErrors.New("select query failed")

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(5, nil)
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, expectedError)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllCouponsByShopIDUseCase_Execute_SubsequentPage(t *testing.T) {
	shopID := 1

	t.Run("when subsequent page succeeds then returns coupons without totalCount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		lastID := 5
		filters := models.CouponFilters{
			LastID: &lastID,
		}
		expectedCoupons := newTestCoupons()

		couponServiceMock := mocks.NewCouponService(t)
		// Only GetAllByShopIDWithFilters should be called, NOT CountByShopIDWithFilters
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCoupons, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", expectedCoupons, models.DefaultLimit, models.SortByCreatedAt).
			Return(expectedCoupons, "", false)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedCoupons, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})

	t.Run("when subsequent page select fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		lastID := 5
		filters := models.CouponFilters{
			LastID: &lastID,
		}
		expectedError := stdErrors.New("database error")

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(nil, expectedError)

		paginationServiceMock := newMockPaginationService(t)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, coupons)
		assert.Empty(t, nextCursor)
		assert.False(t, hasMore)
		assert.Nil(t, totalCount)
	})
}

func TestGetAllCouponsByShopIDUseCase_Execute_Pagination(t *testing.T) {
	t.Run("when has more pages then returns hasMore true and nextCursor", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		shopID := 1
		filters := models.CouponFilters{}
		expectedCount := 50

		// Simulate LIMIT+1 strategy: 3 coupons returned
		allCoupons := []*models.Coupon{
			newTestCouponWithID(1),
			newTestCouponWithID(2),
			newTestCouponWithID(3),
		}
		trimmedCoupons := allCoupons[:2]
		expectedCursor := "eyJpZCI6Miwic29ydF92YWx1ZSI6IjIwMjUtMDEtMDFUMDA6MDA6MDBaIn0="

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			CountByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(expectedCount, nil)
		couponServiceMock.EXPECT().
			GetAllByShopIDWithFilters(mock.Anything, shopID, mock.Anything).
			Return(allCoupons, nil)

		paginationServiceMock := newMockPaginationService(t)
		paginationServiceMock.On("ApplyPagination", allCoupons, models.DefaultLimit, models.SortByCreatedAt).
			Return(trimmedCoupons, expectedCursor, true)

		useCase := NewGetAllCouponsByShopIDUseCase(couponServiceMock, paginationServiceMock)

		// Act
		coupons, nextCursor, hasMore, totalCount, err := useCase.Execute(ctx, shopID, filters)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, trimmedCoupons, coupons)
		assert.Equal(t, expectedCursor, nextCursor)
		assert.True(t, hasMore)
		require.NotNil(t, totalCount)
		assert.Equal(t, expectedCount, *totalCount)
	})
}

// =============================================================================
// ValidateStoreCouponUseCase Tests
// =============================================================================

func TestValidateStoreCouponUseCase_Execute(t *testing.T) {
	t.Run("when store found and coupon valid then returns coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		code := "SAVE10"
		phone := "1234567890"
		subtotal := 100.00
		store := &models.Store{ID: 1, Name: "Test Store", Slug: storeSlug}
		expectedCoupon := newTestCouponWithID(1)
		expectedCoupon.DiscountAmount = 10.00

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			ValidateCoupon(ctx, code, store.ID, phone, subtotal).
			Return(expectedCoupon, nil)

		useCase := NewValidateStoreCouponUseCase(storeServiceMock, couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, storeSlug, code, phone, subtotal)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedCoupon, result)
		assert.Equal(t, 10.00, result.DiscountAmount)
	})

	t.Run("when store not found then returns error without validating coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "non-existent-store"
		code := "SAVE10"
		phone := "1234567890"
		subtotal := 100.00
		notFoundError := &errors.RecordNotFoundError{Message: "store not found"}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(nil, notFoundError)

		couponServiceMock := mocks.NewCouponService(t)
		// ValidateCoupon should NOT be called - testify will fail if it is

		useCase := NewValidateStoreCouponUseCase(storeServiceMock, couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, storeSlug, code, phone, subtotal)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var notFound *errors.RecordNotFoundError
		assert.True(t, stdErrors.As(err, &notFound))
	})

	t.Run("when store found but coupon invalid then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		storeSlug := "test-store"
		code := "EXPIRED"
		phone := "1234567890"
		subtotal := 100.00
		store := &models.Store{ID: 1, Name: "Test Store", Slug: storeSlug}
		validationError := &errors.ValidationError{Message: "coupon has expired"}

		storeServiceMock := mocks.NewStoreService(t)
		storeServiceMock.EXPECT().
			GetBySlug(ctx, storeSlug).
			Return(store, nil)

		couponServiceMock := mocks.NewCouponService(t)
		couponServiceMock.EXPECT().
			ValidateCoupon(ctx, code, store.ID, phone, subtotal).
			Return(nil, validationError)

		useCase := NewValidateStoreCouponUseCase(storeServiceMock, couponServiceMock)

		// Act
		result, err := useCase.Execute(ctx, storeSlug, code, phone, subtotal)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		var valErr *errors.ValidationError
		assert.True(t, stdErrors.As(err, &valErr))
	})
}
