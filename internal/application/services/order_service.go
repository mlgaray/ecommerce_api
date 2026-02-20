package services

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// OrderService contains business logic for order operations.
// Persists orders to the database.
type OrderService struct {
	orderRepository ports.OrderRepository
}

func NewOrderService(orderRepository ports.OrderRepository) *OrderService {
	return &OrderService{
		orderRepository: orderRepository,
	}
}

// Create persists a new order.
// Expects the order to be already validated by StoreService.
// Sets store data from the store parameter.
// Sets status to "pending".
// Calculates totals and validates them against frontend values.
// Validates order business rules before persisting.
func (s *OrderService) Create(ctx context.Context, order *models.Order, store *models.Store) (*models.Order, error) {
	// 1. Set store data in order
	order.Store = &models.Store{
		ID:   store.ID,
		Name: store.Name,
		Slug: store.Slug,
	}

	// 2. Set status to pending
	order.Status = models.OrderStatusPending

	// 3. Calculate totals and validate against frontend values
	if err := s.CalculateAndValidateTotals(order); err != nil {
		return nil, err
	}

	// 4. Validate order business rules
	if err := order.Validate(); err != nil {
		return nil, err
	}

	// 5. Persist order
	return s.orderRepository.Create(ctx, order)
}

// GetAllByShopIDWithFilters retrieves orders with filters.
// Delegates to repository - service adds no extra business logic for reads.
func (s *OrderService) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) ([]*models.Order, error) {
	return s.orderRepository.GetAllByShopIDWithFilters(ctx, shopID, filters)
}

// CountByShopIDWithFilters returns total count of orders matching filters.
// Delegates to repository.
func (s *OrderService) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.OrderFilters) (int, error) {
	return s.orderRepository.CountByShopIDWithFilters(ctx, shopID, filters)
}

// GetByID retrieves a single order with full details.
// Delegates to repository - service adds no extra business logic for reads.
func (s *OrderService) GetByID(ctx context.Context, shopID int, orderID int) (*models.Order, error) {
	return s.orderRepository.GetByID(ctx, shopID, orderID)
}

// UpdateStatus validates the status transition and persists the change.
// Captures current status before mutation for optimistic locking.
func (s *OrderService) UpdateStatus(ctx context.Context, shopID int, order *models.Order, newStatus models.OrderStatus) error {
	currentStatus := order.Status

	if err := order.TransitionTo(newStatus); err != nil {
		return err
	}

	return s.orderRepository.UpdateStatus(ctx, shopID, order.ID, currentStatus, newStatus)
}

// Update persists order changes. Expects the order to be already validated
// by the use case (editability, items, delivery, totals, domain rules).
// Does NOT change the order status.
func (s *OrderService) Update(ctx context.Context, shopID int, order *models.Order) error {
	return s.orderRepository.Update(ctx, shopID, order)
}

// CalculateAndValidateTotals calculates item total prices, subtotal, and total.
// Then validates that the frontend-sent values match the backend calculations.
// Backend is the source of truth for all monetary calculations.
func (s *OrderService) CalculateAndValidateTotals(order *models.Order) error {
	// Calculate item totals and subtotal
	var calculatedSubtotal float64
	for _, item := range order.Items {
		item.TotalPrice = item.UnitPrice * float64(item.Quantity)
		calculatedSubtotal += item.TotalPrice
	}

	// Validate subtotal matches frontend value
	if order.Subtotal != calculatedSubtotal {
		return &errors.ValidationError{Message: errors.SubtotalMismatch}
	}

	// Calculate total
	calculatedTotal := calculatedSubtotal + order.ShippingCost

	// Validate total matches frontend value
	if order.Total != calculatedTotal {
		return &errors.ValidationError{Message: errors.TotalMismatch}
	}

	return nil
}
