package category

import (
	"context"
	"sync"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetAllByShopIDWithFiltersUseCase orchestrates category retrieval with filters and pagination.
// This is the orchestrator layer - coordinates multiple services (NOT repositories directly).
type GetAllByShopIDWithFiltersUseCase struct {
	categoryService   ports.CategoryService
	paginationService ports.PaginationService[*models.Category]
}

// NewGetAllByShopIDWithFiltersUseCase creates a new use case instance.
func NewGetAllByShopIDWithFiltersUseCase(
	categoryService ports.CategoryService,
	paginationService ports.PaginationService[*models.Category],
) ports.GetAllCategoriesByShopIDWithFiltersUseCase {
	return &GetAllByShopIDWithFiltersUseCase{
		categoryService:   categoryService,
		paginationService: paginationService,
	}
}

// Execute orchestrates the complete flow:
//  1. Validates and normalizes filters (once, before parallel execution)
//  2. Fetches total count on first page only (delegates to CategoryService)
//  3. Fetches categories with filters (delegates to CategoryService)
//  4. Applies pagination logic (delegates to PaginationService)
//
// Returns:
//   - categories: List of categories (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - totalCount: Total count of items (only on first page when LastID is nil, otherwise nil)
//   - error: Any error that occurred
func (uc *GetAllByShopIDWithFiltersUseCase) Execute(
	ctx context.Context,
	shopID int,
	filters *models.CategoryFilters,
) ([]*models.Category, string, bool, *int, error) {
	// Validate and normalize filters BEFORE parallel execution (avoids data race)
	if err := filters.Validate(); err != nil {
		return nil, "", false, nil, err
	}

	var (
		totalCount *int
		categories []*models.Category
		err        error
	)

	// First page: execute COUNT and SELECT in parallel to reduce latency
	if filters.LastID == nil {
		var (
			wg               sync.WaitGroup
			count            int
			countErr         error
			categoriesResult []*models.Category
			categoriesErr    error
		)

		// Query 1: COUNT (parallel)
		wg.Go(func() {
			count, countErr = uc.categoryService.CountByShopIDWithFilters(ctx, shopID, *filters)
		})

		// Query 2: SELECT categories (parallel)
		wg.Go(func() {
			categoriesResult, categoriesErr = uc.categoryService.GetAllByShopIDWithFilters(ctx, shopID, *filters)
		})

		wg.Wait()

		// Assign results after both goroutines complete (no race)
		categories = categoriesResult
		err = categoriesErr
		if countErr == nil {
			totalCount = &count
		}
	} else {
		// Subsequent pages: only fetch categories (no count needed)
		categories, err = uc.categoryService.GetAllByShopIDWithFilters(ctx, shopID, *filters)
	}

	if err != nil {
		return nil, "", false, nil, err
	}

	// Apply pagination logic (delegated to PaginationService)
	trimmedCategories, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		categories,
		filters.Limit,
		filters.SortBy,
	)

	return trimmedCategories, nextCursor, hasMore, totalCount, nil
}
