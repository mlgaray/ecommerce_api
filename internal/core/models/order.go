package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// OrderStatus representa los estados posibles de una orden
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// validTransitions defines the allowed state machine transitions.
// Terminal states (completed, cancelled) have no outgoing transitions.
var validTransitions = map[OrderStatus]map[OrderStatus]bool{
	OrderStatusPending:   {OrderStatusConfirmed: true, OrderStatusCancelled: true},
	OrderStatusConfirmed: {OrderStatusCompleted: true, OrderStatusCancelled: true},
}

// CanTransitionTo checks if the order can transition to the given status.
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	allowed, ok := validTransitions[o.Status]
	if !ok {
		return false
	}
	return allowed[newStatus]
}

// TransitionTo validates and applies a status transition.
// Returns ValidationError if the transition is not allowed.
func (o *Order) TransitionTo(newStatus OrderStatus) error {
	if !o.CanTransitionTo(newStatus) {
		return &errors.ValidationError{Message: errors.InvalidOrderTransition}
	}
	o.Status = newStatus
	return nil
}

// IsEditable returns true if the order can be edited.
// Only pending and confirmed orders can be edited.
// Completed and cancelled orders are terminal states and cannot be modified.
func (o *Order) IsEditable() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

// IsValidOrderStatus checks if a string is a valid order status.
func IsValidOrderStatus(s string) bool {
	switch OrderStatus(s) {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusCompleted, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// Order representa una orden del sistema
type Order struct {
	ID             int             `json:"id,omitempty"`
	OrderNumber    string          `json:"order_number,omitempty"`
	Store          *Store          `json:"store,omitempty"`
	Status         OrderStatus     `json:"status,omitempty"`
	Customer       *Customer       `json:"customer,omitempty"`
	PaymentMethod  *PaymentMethod  `json:"payment_method,omitempty"`
	DeliveryMethod *DeliveryMethod `json:"delivery_method,omitempty"`
	Items          []*OrderItem    `json:"items,omitempty"`
	Coupon         *Coupon         `json:"coupon,omitempty"`
	Subtotal       float64         `json:"subtotal"`
	ShippingCost   float64         `json:"shipping_cost"`
	Discount       float64         `json:"discount"`
	Total          float64         `json:"total"`
	CreatedAt      time.Time       `json:"created_at,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`

	// Transient fields (not persisted)
	ItemsCount int `json:"items_count,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (o *Order) GetID() int {
	return o.ID
}

// GetSortValue implements Sortable interface for pagination
func (o *Order) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case OrderSortByTotal:
		return o.Total
	case OrderSortByOrderNumber:
		return o.OrderNumber
	case SortByCreatedAt:
		return o.CreatedAt
	default:
		return nil
	}
}

// RemoveCoupon clears the coupon from the order and recalculates the total.
// Returns false if the order had no coupon (no-op).
func (o *Order) RemoveCoupon() bool {
	if o.Coupon == nil {
		return false
	}
	o.Coupon = nil
	o.Discount = 0
	o.Total = o.Subtotal + o.ShippingCost
	return true
}

// Validate validates business rules for the Order domain model
func (o *Order) Validate() error {
	if o.Customer == nil || o.Customer.Name == "" {
		return &errors.ValidationError{Message: errors.OrderCustomerNameRequired}
	}

	if len(o.Items) == 0 {
		return &errors.ValidationError{Message: errors.OrderMustHaveItems}
	}

	for _, item := range o.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}
