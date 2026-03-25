package staff

import (
	"context"
	"sync"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetAllStaffByShopIDUseCase struct {
	staffService      ports.StaffService
	paginationService ports.PaginationService[*models.Staff]
}

func NewGetAllStaffByShopIDUseCase(
	staffService ports.StaffService,
	paginationService ports.PaginationService[*models.Staff],
) ports.GetAllStaffByShopIDUseCase {
	return &GetAllStaffByShopIDUseCase{
		staffService:      staffService,
		paginationService: paginationService,
	}
}

func (uc *GetAllStaffByShopIDUseCase) Execute(
	ctx context.Context,
	shopID int,
	filters models.StaffFilters,
) ([]*models.Staff, string, bool, *int, error) {
	validatedFilters, err := filters.Validated()
	if err != nil {
		return nil, "", false, nil, err
	}

	var totalCount *int
	var staffList []*models.Staff

	// First page: parallel COUNT + SELECT
	if validatedFilters.LastID == nil {
		var wg sync.WaitGroup
		var count int
		var countErr, listErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			count, countErr = uc.staffService.CountByShopIDWithFilters(ctx, shopID, validatedFilters)
		}()
		go func() {
			defer wg.Done()
			staffList, listErr = uc.staffService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
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
		staffList, err = uc.staffService.GetAllByShopIDWithFilters(ctx, shopID, validatedFilters)
		if err != nil {
			return nil, "", false, nil, err
		}
	}

	trimmedStaff, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		staffList,
		validatedFilters.Limit,
		validatedFilters.SortBy,
	)

	return trimmedStaff, nextCursor, hasMore, totalCount, nil
}
