package category

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// DeleteCategoryUseCase orchestrates the category deletion flow.
// It acts as a pure coordinator, delegating all logic to CategoryService.
type DeleteCategoryUseCase struct {
	categoryService ports.CategoryService
}

// NewDeleteCategoryUseCase creates a new DeleteCategoryUseCase instance.
func NewDeleteCategoryUseCase(categoryService ports.CategoryService) ports.DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute deletes a category by delegating to the CategoryService.
// The service handles shop validation, persistence, and image cleanup.
func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, categoryID, shopID int) error {
	return uc.categoryService.Delete(ctx, categoryID, shopID)
}
