package category

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// UpdateCategoryUseCase orchestrates the category update flow.
// It acts as a pure coordinator, delegating all logic to CategoryService.
type UpdateCategoryUseCase struct {
	categoryService ports.CategoryService
}

// NewUpdateCategoryUseCase creates a new UpdateCategoryUseCase instance.
func NewUpdateCategoryUseCase(categoryService ports.CategoryService) ports.UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute updates a category by delegating to the CategoryService.
// The service handles image upload, persistence, and old image cleanup.
func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, categoryID int, category *models.Category, newImageBuffer []byte, shopID int) error {
	return uc.categoryService.Update(ctx, categoryID, category, newImageBuffer, shopID)
}
