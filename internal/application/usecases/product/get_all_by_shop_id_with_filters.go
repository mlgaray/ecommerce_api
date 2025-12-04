package product

import (
	"context"
	"sync"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetAllByShopIDWithFiltersUseCase orchestrates the product retrieval with filters and pagination
// This is the orchestrator layer - coordinates multiple services (NOT repositories directly)
type GetAllByShopIDWithFiltersUseCase struct {
	productService    ports.ProductService
	paginationService ports.PaginationService[*models.Product]
}

func NewGetAllByShopIDWithFiltersUseCase(
	productService ports.ProductService,
	paginationService ports.PaginationService[*models.Product],
) ports.GetAllByShopIDWithFiltersUseCase {
	return &GetAllByShopIDWithFiltersUseCase{
		productService:    productService,
		paginationService: paginationService,
	}
}

// Execute orchestrates the complete flow:
//  1. Fetches total count on first page only (delegates to ProductService)
//  2. Fetches products with filters (delegates to ProductService - validates internally)
//  3. Applies pagination logic (delegates to PaginationService)
//
// ShopID is a context parameter (not a filter), passed separately.
//
// Returns:
//   - products: List of products (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - totalCount: Total count of items (only on first page when LastID is nil, otherwise nil)
//   - error: Any error that occurred
func (uc *GetAllByShopIDWithFiltersUseCase) Execute(
	ctx context.Context,
	shopID int,
	filters *models.ProductFilters,
) ([]*models.Product, string, bool, *int, error) {
	var (
		totalCount *int
		products   []*models.Product
		err        error
	)

	// First page: execute COUNT and SELECT in parallel to reduce latency
	if filters.LastID == nil {
		var wg sync.WaitGroup

		// Query 1: COUNT (parallel)
		wg.Go(func() {
			count, countErr := uc.productService.CountByShopIDWithFilters(ctx, shopID, *filters)
			if countErr == nil {
				totalCount = &count
			}
			// Log error but don't fail - count is optional
		})

		// Query 2: SELECT products (parallel)
		wg.Go(func() {
			products, err = uc.productService.GetAllByShopIDWithFilters(ctx, shopID, filters)
		})

		wg.Wait()
	} else {
		// Subsequent pages: only fetch products (no count needed)
		products, err = uc.productService.GetAllByShopIDWithFilters(ctx, shopID, filters)
	}

	if err != nil {
		return nil, "", false, nil, err
	}

	// Apply pagination logic (delegated to PaginationService)
	trimmedProducts, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		products,
		filters.Limit,
		filters.SortBy,
	)

	return trimmedProducts, nextCursor, hasMore, totalCount, nil
}
