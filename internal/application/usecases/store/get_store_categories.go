package store

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetStoreCategoriesUseCase orchestrates category retrieval by store slug.
// This is the orchestrator layer - coordinates multiple services.
// Note: This use case does NOT return totalCount as the store frontend doesn't use pagination UI.
type GetStoreCategoriesUseCase struct {
	storeService      ports.StoreService
	categoryService   ports.CategoryService
	paginationService ports.PaginationService[*models.Category]
}

// NewGetStoreCategoriesUseCase creates a new use case instance.
func NewGetStoreCategoriesUseCase(
	storeService ports.StoreService,
	categoryService ports.CategoryService,
	paginationService ports.PaginationService[*models.Category],
) ports.GetStoreCategoriesUseCase {
	return &GetStoreCategoriesUseCase{
		storeService:      storeService,
		categoryService:   categoryService,
		paginationService: paginationService,
	}
}

// Execute orchestrates the complete flow:
//  1. Resolves slug to store (gets shop ID)
//  2. Validates and normalizes filters
//  3. Fetches categories with filters
//  4. Applies pagination logic
//
// Returns:
//   - categories: List of categories (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - error: Any error that occurred
func (uc *GetStoreCategoriesUseCase) Execute(
	ctx context.Context,
	slug string,
	filters models.CategoryFilters,
) ([]*models.Category, string, bool, error) {
	// 1. Resolve slug to store (get shop ID)
	store, err := uc.storeService.GetBySlug(ctx, slug)
	if err != nil {
		return nil, "", false, err
	}

	// 2. Validate and normalize filters (immutable - returns a validated copy)
	validatedFilters, err := filters.Validated()
	if err != nil {
		return nil, "", false, err
	}

	// 3. Fetch categories
	categories, err := uc.categoryService.GetAllByShopIDWithFilters(ctx, store.ID, validatedFilters)
	if err != nil {
		return nil, "", false, err
	}

	// 4. Apply pagination logic (delegated to PaginationService)
	trimmedCategories, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		categories,
		validatedFilters.Limit,
		validatedFilters.SortBy,
	)

	return trimmedCategories, nextCursor, hasMore, nil
}
