package product

import (
	"context"

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
// Returns:
//   - products: List of products (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - totalCount: Total count of items (only on first page when LastID is nil, otherwise nil)
//   - error: Any error that occurred
func (uc *GetAllByShopIDWithFiltersUseCase) Execute(
	ctx context.Context,
	filters *models.ProductFilters,
) ([]*models.Product, string, bool, *int, error) {
	// Step 1: Get total count ONLY on first page (when LastID is nil)
	// This allows frontend to show "20 of 1000" without recalculating on every page
	var totalCount *int
	if filters.LastID == nil {
		count, err := uc.productService.CountByShopIDWithFilters(ctx, *filters)
		if err != nil {
			// Log error but don't fail the request - count is optional
			// Frontend can still function without total count
			totalCount = nil
		} else {
			totalCount = &count
		}
	}

	// Step 2: Fetch products with LIMIT+1 strategy
	// Service validates filters internally (normalizes Limit, SortBy, SortOrder)
	// Since filters is a pointer, normalized values propagate back for Step 3
	products, err := uc.productService.GetAllByShopIDWithFilters(ctx, filters)
	if err != nil {
		return nil, "", false, nil, err
	}

	// Step 3: Apply pagination logic (delegated to PaginationService)
	// Uses normalized filters.Limit from Step 2 validation
	// This handles:
	// - Detecting if there are more pages (hasMore)
	// - Building the cursor for the next page
	// - Trimming the extra item (if it exists)
	trimmedProducts, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		products,
		filters.Limit,
		filters.SortBy,
	)

	return trimmedProducts, nextCursor, hasMore, totalCount, nil
}
