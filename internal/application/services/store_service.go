package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// StoreService contains business logic for public store operations.
// Maps Shop data to Store model for customer-facing endpoints.
// Used by CreateOrderUseCase for order validation and creation.
type StoreService struct {
	shopRepository    ports.ShopRepository
	productRepository ports.ProductRepository
}

func NewStoreService(shopRepository ports.ShopRepository, productRepository ports.ProductRepository) *StoreService {
	return &StoreService{
		shopRepository:    shopRepository,
		productRepository: productRepository,
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

// ValidateOrderItems validates that order items match current database data.
// For each item:
// - Validates product exists and is active
// - Compares product data (name, price, promotional status) with frontend data
// - Compares variant/option data (name, price) with frontend data
// - Validates unit price matches calculated price (base + options)
// - Validates stock availability if product is stockeable
//
// Performance: Uses batch query (GetByIDsAndShopID) to fetch all products in 1 query
// instead of N sequential queries.
//
//nolint:gocyclo // Complexity is intentional - sequential validation checks per item (exists, active, data match, price, stock)
func (s *StoreService) ValidateOrderItems(ctx context.Context, items []*models.OrderItem, storeID int) error {
	// 1. Extract and validate product IDs from items
	productIDs := make([]int, 0, len(items))
	for _, item := range items {
		if item.Product == nil || item.Product.ID == 0 {
			return &errors.ValidationError{Message: errors.OrderItemProductRequired}
		}
		productIDs = append(productIDs, item.Product.ID)
	}

	// 2. Batch fetch all products in one query
	productsMap, err := s.productRepository.GetByIDsAndShopID(ctx, productIDs, storeID)
	if err != nil {
		return err
	}

	// 3. Validate each item (O(1) lookup in map)
	for _, item := range items {
		dbProduct, exists := productsMap[item.Product.ID]
		if !exists {
			return &errors.RecordNotFoundError{Message: errors.ProductNotFound}
		}

		// Validate product is active
		if !dbProduct.IsActive {
			return &errors.ValidationError{Message: errors.ProductNotActive}
		}

		// Validate product data matches (name, price, promotional)
		if !dbProduct.EqualsForOrder(item.Product) {
			return &errors.ValidationError{Message: errors.ProductDataMismatch}
		}

		// Validate variants/options match (if any selected)
		if len(item.Product.Variants) > 0 {
			if !dbProduct.VariantsEqualForOrder(item.Product.Variants) {
				return &errors.ValidationError{Message: errors.ProductDataMismatch}
			}
		}

		// Validate unit price matches calculated price
		if err := s.validateUnitPrice(item, dbProduct); err != nil {
			return err
		}

		// Validate stock availability
		if !dbProduct.HasSufficientStock(item.Quantity) {
			return &errors.BusinessRuleError{Message: errors.InsufficientStock}
		}
	}

	return nil
}

// validateUnitPrice validates that the item's unit price matches the calculated price.
// Calculated price = effective base price + sum of (option price * option quantity).
// Effective base price is promotional_price if promotional, otherwise regular price.
func (s *StoreService) validateUnitPrice(item *models.OrderItem, dbProduct *models.Product) error {
	// Calculate expected unit price
	expectedUnitPrice := dbProduct.GetEffectivePrice()

	// Add option prices from selected variants (price * quantity)
	for _, variant := range item.Product.Variants {
		for _, option := range variant.Options {
			quantity := option.Quantity
			if quantity <= 0 {
				quantity = 1
			}
			expectedUnitPrice += option.Price * float64(quantity)
		}
	}

	// Compare with item's unit price
	if item.UnitPrice != expectedUnitPrice {
		return &errors.ValidationError{Message: errors.UnitPriceMismatch}
	}

	return nil
}

// ValidateDeliveryMethod validates that the delivery method is active and shipping cost matches.
// Looks up delivery method by ID in store's delivery methods.
// Validates the method is active.
// Validates shipping cost matches the delivery configuration (fixed price, zone-based, or pickup).
func (s *StoreService) ValidateDeliveryMethod(store *models.Store, deliveryMethod *models.DeliveryMethod, shippingCost float64) error {
	if deliveryMethod == nil || deliveryMethod.ID == 0 {
		return &errors.ValidationError{Message: errors.OrderDeliveryMethodRequired}
	}

	// Find delivery method in store's delivery methods
	var dbMethod *models.DeliveryMethod
	for _, dm := range store.DeliveryMethods {
		if dm.ID == deliveryMethod.ID {
			dbMethod = dm
			break
		}
	}

	if dbMethod == nil {
		return &errors.ValidationError{Message: errors.DeliveryMethodNotFound}
	}

	if !dbMethod.IsActive {
		return &errors.ValidationError{Message: errors.DeliveryMethodNotFound}
	}

	// Validate shipping cost matches delivery configuration
	selectedZoneID := s.getSelectedDeliveryZoneID(deliveryMethod)
	return s.validateShippingCost(shippingCost, dbMethod, selectedZoneID)
}

// getSelectedDeliveryZoneID extracts the selected delivery zone ID from the delivery method.
func (s *StoreService) getSelectedDeliveryZoneID(deliveryMethod *models.DeliveryMethod) int {
	if deliveryMethod != nil &&
		len(deliveryMethod.DeliveryZones) > 0 &&
		deliveryMethod.DeliveryZones[0] != nil {
		return deliveryMethod.DeliveryZones[0].ID
	}
	return 0
}

// validateShippingCost validates the shipping cost matches the delivery method configuration.
//
//nolint:gocyclo // Complexity is intentional - sequential validation for each delivery type (pickup, fixed, zones, free)
func (s *StoreService) validateShippingCost(shippingCost float64, method *models.DeliveryMethod, selectedZoneID int) error {
	// Pickup should have zero shipping cost
	if method.Code == models.DeliveryMethodPickup {
		if shippingCost != 0 {
			return &errors.ValidationError{Message: errors.PickupShouldHaveZeroCost}
		}
		return nil
	}

	// Delivery method - check for fixed price first
	if method.DeliveryConfig != nil && method.DeliveryConfig.FixedPrice != nil {
		if shippingCost != *method.DeliveryConfig.FixedPrice {
			return &errors.ValidationError{Message: errors.ShippingCostMismatch}
		}
		return nil
	}

	// Zone-based pricing
	if len(method.DeliveryZones) > 0 {
		if selectedZoneID == 0 {
			return &errors.ValidationError{Message: errors.DeliveryZoneRequired}
		}

		zone := s.findDeliveryZone(method.DeliveryZones, selectedZoneID)
		if zone == nil {
			return &errors.ValidationError{Message: errors.DeliveryZoneNotFound}
		}

		if shippingCost != zone.Price {
			return &errors.ValidationError{Message: errors.ShippingCostMismatch}
		}
		return nil
	}

	// No configuration - shipping cost should be zero (free delivery)
	if shippingCost != 0 {
		return &errors.ValidationError{Message: errors.ShippingCostMismatch}
	}

	return nil
}

// findDeliveryZone searches for a delivery zone by ID in the method's zones.
func (s *StoreService) findDeliveryZone(zones []*models.DeliveryZone, id int) *models.DeliveryZone {
	for _, zone := range zones {
		if zone.ID == id {
			return zone
		}
	}
	return nil
}
