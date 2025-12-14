package product

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// DeleteProductUseCase orchestrates the product deletion flow.
// It acts as a pure coordinator, delegating all logic to ProductService.
type DeleteProductUseCase struct {
	productService ports.ProductService
}

// NewDeleteProductUseCase creates a new DeleteProductUseCase instance.
func NewDeleteProductUseCase(productService ports.ProductService) ports.DeleteProductUseCase {
	return &DeleteProductUseCase{
		productService: productService,
	}
}

// Execute deletes a product by delegating to the ProductService.
// The service handles shop validation, persistence, and image cleanup.
func (uc *DeleteProductUseCase) Execute(ctx context.Context, productID, shopID int) error {
	return uc.productService.Delete(ctx, productID, shopID)
}
