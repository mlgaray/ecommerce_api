package store

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type GetStoreBySlugUseCase struct {
	storeService ports.StoreService
}

func NewGetStoreBySlugUseCase(storeService ports.StoreService) ports.GetStoreBySlugUseCase {
	return &GetStoreBySlugUseCase{
		storeService: storeService,
	}
}

func (uc *GetStoreBySlugUseCase) Execute(ctx context.Context, slug string) (*models.Store, error) {
	return uc.storeService.GetBySlug(ctx, slug)
}
