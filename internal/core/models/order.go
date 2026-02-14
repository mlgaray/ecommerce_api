package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// OrderStatus representa los estados posibles de una orden
type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

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
	Subtotal       float64         `json:"subtotal"`
	ShippingCost   float64         `json:"shipping_cost"`
	Total          float64         `json:"total"`
	CreatedAt      time.Time       `json:"created_at,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`

	// Transient field for list views (not persisted)
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
