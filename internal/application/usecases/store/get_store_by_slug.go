package store

import (
	"context"
	"time"

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
	store, err := uc.storeService.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Fire-and-forget: record visit asynchronously so it doesn't delay the response.
	// Uses a timeout context to avoid leaking goroutines if the DB is slow/hung.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = uc.storeService.RecordVisit(bgCtx, store.ID)
	}()

	return store, nil
}
