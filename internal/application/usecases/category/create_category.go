package category

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CreateCategoryUseCase orchestrates the category creation flow.
// It acts as a pure coordinator, delegating all logic to CategoryService.
type CreateCategoryUseCase struct {
	categoryService ports.CategoryService
}

// NewCreateCategoryUseCase creates a new CreateCategoryUseCase instance.
func NewCreateCategoryUseCase(categoryService ports.CategoryService) ports.CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute creates a new category by delegating to the CategoryService.
// The service handles validation, image upload, and persistence.
func (uc *CreateCategoryUseCase) Execute(ctx context.Context, category *models.Category, imageBuffer []byte, shopID int) (*models.Category, error) {
	return uc.categoryService.Create(ctx, category, imageBuffer, shopID)
}
