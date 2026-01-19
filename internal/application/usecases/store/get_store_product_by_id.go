package store

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetStoreProductByIDUseCase orchestrates product retrieval by ID within a store.
// This is the orchestrator layer - coordinates multiple services.
// Returns product only if it exists, belongs to the store, and is active.
type GetStoreProductByIDUseCase struct {
	storeService   ports.StoreService
	productService ports.ProductService
}

// NewGetStoreProductByIDUseCase creates a new use case instance.
func NewGetStoreProductByIDUseCase(
	storeService ports.StoreService,
	productService ports.ProductService,
) ports.GetStoreProductByIDUseCase {
	return &GetStoreProductByIDUseCase{
		storeService:   storeService,
		productService: productService,
	}
}

// Execute orchestrates the complete flow:
//  1. Resolves slug to store (gets shop ID)
//  2. Fetches product by ID and validates shop ownership
//  3. Validates product is active (public storefront only shows active products)
//
// Returns:
//   - product: The product with variants and options
//   - error: store_not_found, product_not_found, or other errors
func (uc *GetStoreProductByIDUseCase) Execute(
	ctx context.Context,
	slug string,
	productID int,
) (*models.Product, error) {
	// 1. Resolve slug to store (get shop ID)
	store, err := uc.storeService.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// 2. Fetch product by ID and validate shop ownership
	product, err := uc.productService.GetByIDAndShopID(ctx, productID, store.ID)
	if err != nil {
		return nil, err
	}

	// 3. Validate product is active (public storefront only shows active products)
	if !product.IsActive {
		return nil, &errors.RecordNotFoundError{Message: errors.ProductNotFound}
	}

	return product, nil
}
