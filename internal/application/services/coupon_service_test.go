package services

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestCoupon() *models.Coupon {
	return &models.Coupon{
		ID:       1,
		ShopID:   1,
		Code:     "SUMMER20",
		Type:     models.CouponTypePercentage,
		Value:    20,
		IsActive: true,
	}
}

func ptrInt(v int) *int              { return &v }
func ptrTime(v time.Time) *time.Time { return &v }

// =============================================================================
// CouponService.Create Tests
// =============================================================================

func TestCouponService_Create(t *testing.T) {
	t.Run("when coupon is valid then creates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := &models.Coupon{
			Code:  "summer20",
			Type:  models.CouponTypePercentage,
			Value: 20,
		}
		expectedCoupon := &models.Coupon{
			ID:     1,
			ShopID: 5,
			Code:   "SUMMER20",
			Type:   models.CouponTypePercentage,
			Value:  20,
		}

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			Create(ctx, coupon).
			Return(expectedCoupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.Create(ctx, coupon, 5)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCoupon, result)
		assert.Equal(t, "SUMMER20", coupon.Code)
		assert.Equal(t, 5, coupon.ShopID)
	})

	t.Run("when coupon code is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := &models.Coupon{
			Code:  "",
			Type:  models.CouponTypePercentage,
			Value: 20,
		}

		repoMock := mocks.NewCouponRepository(t)
		service := NewCouponService(repoMock)

		// Act
		result, err := service.Create(ctx, coupon, 5)

		// Assert
		assert.Nil(t, result)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponCodeIsRequired, validationErr.Message)
	})

	t.Run("when code has whitespace then trims and uppercases", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := &models.Coupon{
			Code:  "  promo10  ",
			Type:  models.CouponTypeFixed,
			Value: 100,
		}

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			Create(ctx, coupon).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		_, err := service.Create(ctx, coupon, 1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "PROMO10", coupon.Code)
	})
}

// =============================================================================
// CouponService.Update Tests
// =============================================================================

func TestCouponService_Update(t *testing.T) {
	t.Run("when coupon is valid then updates successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := &models.Coupon{
			Code:  "winter30",
			Type:  models.CouponTypePercentage,
			Value: 30,
		}

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			Update(ctx, coupon).
			Return(nil)

		service := NewCouponService(repoMock)

		// Act
		err := service.Update(ctx, 1, 5, coupon)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, coupon.ID)
		assert.Equal(t, 5, coupon.ShopID)
		assert.Equal(t, "WINTER30", coupon.Code)
	})

	t.Run("when coupon is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := &models.Coupon{
			Code:  "TEST",
			Type:  "invalid",
			Value: 10,
		}

		repoMock := mocks.NewCouponRepository(t)
		service := NewCouponService(repoMock)

		// Act
		err := service.Update(ctx, 1, 5, coupon)

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponTypeInvalid, validationErr.Message)
	})
}

// =============================================================================
// CouponService.ValidateCoupon Tests
// =============================================================================

func TestCouponService_ValidateCoupon(t *testing.T) {
	t.Run("when coupon is valid then returns coupon with discount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200.0, result.DiscountAmount)
	})

	t.Run("when code has lowercase and spaces then normalizes before lookup", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()

		repoMock := mocks.NewCouponRepository(t)
		// Expect the normalized (uppercase, trimmed) code
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act — pass lowercase with spaces
		result, err := service.ValidateCoupon(ctx, "  summer20  ", 1, "", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when coupon not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "INVALID", 1).
			Return(nil, &errors.RecordNotFoundError{Message: errors.CouponNotFound})

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "INVALID", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var notFoundErr *errors.RecordNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("when coupon is not active then returns validation error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.IsActive = false

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponNotActive, bizErr.Message)
	})

	t.Run("when coupon has expired then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.ExpiresAt = ptrTime(time.Now().Add(-24 * time.Hour))

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponExpired, bizErr.Message)
	})

	t.Run("when coupon has not started yet then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.StartsAt = ptrTime(time.Now().Add(24 * time.Hour))

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponNotYetStarted, bizErr.Message)
	})

	t.Run("when coupon is within validity window then succeeds", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.StartsAt = ptrTime(time.Now().Add(-1 * time.Hour))
		coupon.ExpiresAt = ptrTime(time.Now().Add(24 * time.Hour))

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200.0, result.DiscountAmount)
	})

	t.Run("when subtotal is below minimum then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.MinOrderAmount = 500

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 100)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponMinOrderNotMet, bizErr.Message)
	})

	t.Run("when usage limit reached then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.UsageLimit = ptrInt(10)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsages(ctx, 1).
			Return(10, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponUsageLimitReached, bizErr.Message)
	})

	t.Run("when usage limit not reached then continues validation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.UsageLimit = ptrInt(10)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsages(ctx, 1).
			Return(5, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when count usages fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.UsageLimit = ptrInt(10)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsages(ctx, 1).
			Return(0, stdErrors.New("db error"))

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		assert.EqualError(t, err, "db error")
	})

	t.Run("when phone usage limit reached then returns business rule error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.MaxUsesPerPhone = ptrInt(2)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsagesByPhone(ctx, 1, "1234567890").
			Return(2, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		var bizErr *errors.BusinessRuleError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, errors.CouponPhoneLimitReached, bizErr.Message)
	})

	t.Run("when phone is empty then skips phone limit check", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.MaxUsesPerPhone = ptrInt(1)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("when count usages by phone fails then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.MaxUsesPerPhone = ptrInt(5)

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsagesByPhone(ctx, 1, "1234567890").
			Return(0, stdErrors.New("db error"))

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.Nil(t, result)
		assert.EqualError(t, err, "db error")
	})

	t.Run("when fixed coupon then calculates correct discount", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.Type = models.CouponTypeFixed
		coupon.Value = 150

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "", 1000)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 150.0, result.DiscountAmount)
	})

	t.Run("when all validations pass with limits then returns coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		coupon := newTestCoupon()
		coupon.MinOrderAmount = 500
		coupon.UsageLimit = ptrInt(100)
		coupon.MaxUsesPerPhone = ptrInt(3)
		coupon.StartsAt = ptrTime(time.Now().Add(-1 * time.Hour))
		coupon.ExpiresAt = ptrTime(time.Now().Add(24 * time.Hour))

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByCodeAndShopID(ctx, "SUMMER20", 1).
			Return(coupon, nil)
		repoMock.EXPECT().
			CountUsages(ctx, 1).
			Return(50, nil)
		repoMock.EXPECT().
			CountUsagesByPhone(ctx, 1, "1234567890").
			Return(1, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.ValidateCoupon(ctx, "SUMMER20", 1, "1234567890", 1000)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200.0, result.DiscountAmount)
	})
}

// =============================================================================
// CouponService.Delete Tests
// =============================================================================

func TestCouponService_Delete(t *testing.T) {
	t.Run("when coupon exists then deletes successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			Delete(ctx, 1, 5).
			Return(nil)

		service := NewCouponService(repoMock)

		// Act
		err := service.Delete(ctx, 1, 5)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when coupon not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			Delete(ctx, 99, 5).
			Return(&errors.RecordNotFoundError{Message: errors.CouponNotFound})

		service := NewCouponService(repoMock)

		// Act
		err := service.Delete(ctx, 99, 5)

		// Assert
		var notFoundErr *errors.RecordNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}

// =============================================================================
// CouponService.GetByID Tests
// =============================================================================

func TestCouponService_GetByID(t *testing.T) {
	t.Run("when coupon exists then returns coupon", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		expected := newTestCoupon()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, 1, 1).
			Return(expected, nil)

		service := NewCouponService(repoMock)

		// Act
		result, err := service.GetByID(ctx, 1, 1)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("when coupon not found then returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		repoMock := mocks.NewCouponRepository(t)
		repoMock.EXPECT().
			GetByID(ctx, 99, 1).
			Return(nil, &errors.RecordNotFoundError{Message: errors.CouponNotFound})

		service := NewCouponService(repoMock)

		// Act
		result, err := service.GetByID(ctx, 99, 1)

		// Assert
		assert.Nil(t, result)
		var notFoundErr *errors.RecordNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})
}
