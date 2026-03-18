package store

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

type CheckSlugAvailabilityUseCase struct {
	storeService ports.StoreService
}

func NewCheckSlugAvailabilityUseCase(storeService ports.StoreService) ports.CheckSlugAvailabilityUseCase {
	return &CheckSlugAvailabilityUseCase{
		storeService: storeService,
	}
}

func (uc *CheckSlugAvailabilityUseCase) Execute(ctx context.Context, slug string) (bool, error) {
	return uc.storeService.CheckSlugAvailability(ctx, slug)
}
