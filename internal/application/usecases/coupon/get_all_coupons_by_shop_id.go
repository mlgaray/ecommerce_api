package coupon

import (
	"context"
	"sync"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetAllCouponsByShopIDUseCase orchestrates retrieving coupons with filters and pagination.
type GetAllCouponsByShopIDUseCase struct {
	couponService     ports.CouponService
	paginationService ports.PaginationService[*models.Coupon]
}

func NewGetAllCouponsByShopIDUseCase(
	couponService ports.CouponService,
	paginationService ports.PaginationService[*models.Coupon],
) ports.GetAllCouponsByShopIDUseCase {
	return &GetAllCouponsByShopIDUseCase{
		couponService:     couponService,
		paginationService: paginationService,
	}
}

func (uc *GetAllCouponsByShopIDUseCase) Execute(
	ctx context.Context,
	shopID int,
	filters models.CouponFilters,
) ([]*models.Coupon, string, bool, *int, error) {
	// 1. Validate and normalize filters (immutable copy)
	validatedFilters, err := filters.Validated()
	if err != nil {
		return nil, "", false, nil, err
	}

	var totalCount *int
	var coupons []*models.Coupon

	// 2. First page: parallel COUNT + SELECT
	if validatedFilters.LastID == nil {
		var wg sync.WaitGroup
		var count int
		var countErr, listErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			count, countErr = uc.couponService.CountByShopIDWithFilters(ctx, shopID, validatedFilters)
		}()
		go func() {
			defer wg.Done()
			coupons, listErr = uc.couponService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
		}()
		wg.Wait()

		if countErr != nil {
			return nil, "", false, nil, countErr
		}
		if listErr != nil {
			return nil, "", false, nil, listErr
		}
		totalCount = &count
	} else {
		// Subsequent pages: only SELECT
		coupons, err = uc.couponService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
		if err != nil {
			return nil, "", false, nil, err
		}
	}

	// 3. Apply pagination (LIMIT+1 strategy: trim extra item, build cursor)
	trimmedCoupons, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		coupons,
		validatedFilters.Limit,
		validatedFilters.SortBy,
	)

	return trimmedCoupons, nextCursor, hasMore, totalCount, nil
}
