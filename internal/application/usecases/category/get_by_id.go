package category

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetByIDUseCase struct {
	categoryService ports.CategoryService
}

func NewGetByIDUseCase(categoryService ports.CategoryService) ports.GetCategoryByIDUseCase {
	return &GetByIDUseCase{
		categoryService: categoryService,
	}
}

func (uc *GetByIDUseCase) Execute(ctx context.Context, categoryID int) (*models.Category, error) {
	return uc.categoryService.GetByID(ctx, categoryID)
}
