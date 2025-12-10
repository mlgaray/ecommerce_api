package ports

import "context"

// DeleteCategoryUseCase defines the interface for deleting a category.
type DeleteCategoryUseCase interface {
	// Execute deletes a category by ID.
	// Returns RecordNotFoundError if category doesn't exist or doesn't belong to shop.
	// Returns ReferentialIntegrityError if category has products assigned.
	Execute(ctx context.Context, categoryID, shopID int) error
}
