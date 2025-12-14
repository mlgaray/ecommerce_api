package ports

import "context"

// DeleteProductUseCase defines the interface for deleting a product.
type DeleteProductUseCase interface {
	// Execute deletes a product by ID.
	// Returns RecordNotFoundError if product doesn't exist or doesn't belong to shop.
	// Related entities (images, variants, variant_options) are cascade-deleted automatically.
	Execute(ctx context.Context, productID, shopID int) error
}
