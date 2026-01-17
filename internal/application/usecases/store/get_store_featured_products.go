package store

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetStoreFeaturedProductsUseCase orchestrates featured products retrieval by store slug.
// This is the orchestrator layer - coordinates multiple services.
// Automatically applies is_highlighted=true and is_active=true filters.
// Note: This use case does NOT return totalCount as the store frontend doesn't need it.
type GetStoreFeaturedProductsUseCase struct {
	storeService      ports.StoreService
	productService    ports.ProductService
	paginationService ports.PaginationService[*models.Product]
}

// NewGetStoreFeaturedProductsUseCase creates a new use case instance.
func NewGetStoreFeaturedProductsUseCase(
	storeService ports.StoreService,
	productService ports.ProductService,
	paginationService ports.PaginationService[*models.Product],
) ports.GetStoreFeaturedProductsUseCase {
	return &GetStoreFeaturedProductsUseCase{
		storeService:      storeService,
		productService:    productService,
		paginationService: paginationService,
	}
}

// Execute orchestrates the complete flow:
//  1. Resolves slug to store (gets shop ID)
//  2. Applies featured filters (is_highlighted=true, is_active=true)
//  3. Validates and normalizes filters
//  4. Fetches products with filters
//  5. Applies pagination logic
//
// Returns:
//   - products: List of featured products (max of 'limit' items)
//   - nextCursor: Opaque cursor for next page (empty if no more pages)
//   - hasMore: true if there are more pages
//   - error: Any error that occurred
func (uc *GetStoreFeaturedProductsUseCase) Execute(
	ctx context.Context,
	slug string,
	filters models.ProductFilters,
) ([]*models.Product, string, bool, error) {
	// 1. Resolve slug to store (get shop ID)
	store, err := uc.storeService.GetBySlug(ctx, slug)
	if err != nil {
		return nil, "", false, err
	}

	// 2. Apply featured filters (forced - these define "featured")
	isHighlighted := true
	isActive := true
	filters.IsHighlighted = &isHighlighted
	filters.IsActive = &isActive

	// 3. Validate and normalize filters (immutable - returns a validated copy)
	validatedFilters, err := filters.Validated()
	if err != nil {
		return nil, "", false, err
	}

	// 4. Fetch products
	products, err := uc.productService.GetAllByShopIDWithFilters(ctx, store.ID, validatedFilters)
	if err != nil {
		return nil, "", false, err
	}

	// 5. Apply pagination logic (delegated to PaginationService)
	trimmedProducts, nextCursor, hasMore := uc.paginationService.ApplyPagination(
		products,
		validatedFilters.Limit,
		validatedFilters.SortBy,
	)

	return trimmedProducts, nextCursor, hasMore, nil
}
