package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// StoreService contains business logic for public store operations.
// Maps Shop data to Store model for customer-facing endpoints.
type StoreService struct {
	shopRepository ports.ShopRepository
}

func NewStoreService(shopRepository ports.ShopRepository) *StoreService {
	return &StoreService{
		shopRepository: shopRepository,
	}
}

// GetBySlug retrieves a store by its slug.
// Fetches shop data from repository and maps it to Store model.
func (s *StoreService) GetBySlug(ctx context.Context, slug string) (*models.Store, error) {
	// 1. Get shop from repository
	shop, err := s.shopRepository.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// 2. Map Shop to Store
	store := models.NewStoreFromShop(shop)

	// Future: Calculate additional fields like IsOpen based on operating schedules

	return store, nil
}
