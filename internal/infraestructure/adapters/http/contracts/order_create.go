package contracts

import (
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// CreateOrderRequest represents the HTTP request for creating an order.
// Wraps OrderRequest for extensibility.
type CreateOrderRequest struct {
	Order OrderRequest `json:"order"`
}

// OrderRequest represents the order data in the create request.
// Structure mirrors the domain model for cohesion.
type OrderRequest struct {
	Customer       OrderCustomerRequest       `json:"customer"`
	PaymentMethod  OrderPaymentMethodRequest  `json:"payment_method"`
	DeliveryMethod OrderDeliveryMethodRequest `json:"delivery_method"`
	Items          []OrderItemRequest         `json:"items"`
	Subtotal       float64                    `json:"subtotal"`
	Total          float64                    `json:"total"`
}

// OrderCustomerRequest represents customer data in the order request.
type OrderCustomerRequest struct {
	Name    string               `json:"name"`
	Phone   string               `json:"phone,omitempty"`
	Email   string               `json:"email,omitempty"`
	Address *OrderAddressRequest `json:"address,omitempty"`
}

// OrderAddressRequest represents address data in the customer request.
type OrderAddressRequest struct {
	Name    string  `json:"name"`
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}

// OrderPaymentMethodRequest represents payment method selection in the order request.
// Includes snapshot fields for historical preservation.
type OrderPaymentMethodRequest struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// OrderDeliveryMethodRequest represents delivery method selection in the order request.
// ShippingCost is calculated by frontend and validated by backend.
// DeliveryZone is required when the delivery method uses zone-based pricing.
// Includes snapshot fields for historical preservation.
type OrderDeliveryMethodRequest struct {
	ID           int                       `json:"id"`
	Name         string                    `json:"name"`
	Code         string                    `json:"code"`
	DeliveryZone *OrderDeliveryZoneRequest `json:"delivery_zone,omitempty"`
	ShippingCost float64                   `json:"shipping_cost"`
}

// OrderDeliveryZoneRequest represents a selected delivery zone in the order request.
// Includes snapshot fields for historical preservation.
type OrderDeliveryZoneRequest struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// OrderItemRequest represents an item in the order request.
type OrderItemRequest struct {
	Product         OrderProductRequest          `json:"product"`
	Quantity        int                          `json:"quantity"`
	UnitPrice       float64                      `json:"unit_price"`
	SelectedOptions []OrderSelectedOptionRequest `json:"selected_options,omitempty"`
}

// OrderProductRequest represents product selection in an order item request.
// Includes snapshot fields for historical preservation.
type OrderProductRequest struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	ImageURL         string  `json:"image_url,omitempty"`
	Price            float64 `json:"price"`
	IsPromotional    bool    `json:"is_promotional,omitempty"`
	PromotionalPrice float64 `json:"promotional_price,omitempty"`
}

// OrderSelectedOptionRequest represents a selected variant option in an order item request.
// Includes snapshot fields for historical preservation.
type OrderSelectedOptionRequest struct {
	VariantID   int     `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	OptionID    int     `json:"option_id"`
	OptionName  string  `json:"option_name"`
	OptionPrice float64 `json:"option_price"`
}

// Validate validates HTTP input for creating an order.
func (r *CreateOrderRequest) Validate() error {
	return r.Order.Validate()
}

// Validate validates the order data.
func (r *OrderRequest) Validate() error {
	// Validate customer
	if err := r.validateCustomer(); err != nil {
		return err
	}

	// Validate payment method
	if r.PaymentMethod.ID <= 0 {
		return &httpErrors.BadRequestError{Message: "payment_method_id_is_required"}
	}
	if strings.TrimSpace(r.PaymentMethod.Name) == "" {
		return &httpErrors.BadRequestError{Message: "payment_method_name_is_required"}
	}
	if strings.TrimSpace(r.PaymentMethod.Code) == "" {
		return &httpErrors.BadRequestError{Message: "payment_method_code_is_required"}
	}

	// Validate delivery method
	if err := r.validateDeliveryMethod(); err != nil {
		return err
	}

	// Validate items
	if err := r.validateItems(); err != nil {
		return err
	}

	return nil
}

// validateCustomer validates customer data.
func (r *OrderRequest) validateCustomer() error {
	if strings.TrimSpace(r.Customer.Name) == "" {
		return &httpErrors.BadRequestError{Message: "customer_name_is_required"}
	}
	return nil
}

// validateDeliveryMethod validates delivery method data.
func (r *OrderRequest) validateDeliveryMethod() error {
	if r.DeliveryMethod.ID <= 0 {
		return &httpErrors.BadRequestError{Message: "delivery_method_id_is_required"}
	}
	if strings.TrimSpace(r.DeliveryMethod.Name) == "" {
		return &httpErrors.BadRequestError{Message: "delivery_method_name_is_required"}
	}
	if strings.TrimSpace(r.DeliveryMethod.Code) == "" {
		return &httpErrors.BadRequestError{Message: "delivery_method_code_is_required"}
	}

	if r.DeliveryMethod.ShippingCost < 0 {
		return &httpErrors.BadRequestError{Message: "shipping_cost_cannot_be_negative"}
	}

	// If delivery zone is provided, validate it
	if r.DeliveryMethod.DeliveryZone != nil {
		if r.DeliveryMethod.DeliveryZone.ID <= 0 {
			return &httpErrors.BadRequestError{Message: "delivery_zone_id_is_required"}
		}
		if strings.TrimSpace(r.DeliveryMethod.DeliveryZone.Name) == "" {
			return &httpErrors.BadRequestError{Message: "delivery_zone_name_is_required"}
		}
	}

	return nil
}

// validateItems validates order items.
//
//nolint:gocyclo // Complexity is intentional for readability - linear field validations per item and selected option
func (r *OrderRequest) validateItems() error {
	if len(r.Items) == 0 {
		return &httpErrors.BadRequestError{Message: "order_must_have_at_least_one_item"}
	}

	for _, item := range r.Items {
		if item.Product.ID <= 0 {
			return &httpErrors.BadRequestError{Message: "product_id_is_required"}
		}
		if strings.TrimSpace(item.Product.Name) == "" {
			return &httpErrors.BadRequestError{Message: "product_name_is_required"}
		}
		if item.Product.Price < 0 {
			return &httpErrors.BadRequestError{Message: "product_price_cannot_be_negative"}
		}
		if item.Quantity <= 0 {
			return &httpErrors.BadRequestError{Message: "quantity_must_be_positive"}
		}
		if item.UnitPrice < 0 {
			return &httpErrors.BadRequestError{Message: "unit_price_cannot_be_negative"}
		}

		// Validate selected options
		for _, opt := range item.SelectedOptions {
			if opt.VariantID <= 0 {
				return &httpErrors.BadRequestError{Message: "variant_id_is_required"}
			}
			if strings.TrimSpace(opt.VariantName) == "" {
				return &httpErrors.BadRequestError{Message: "variant_name_is_required"}
			}
			if opt.OptionID <= 0 {
				return &httpErrors.BadRequestError{Message: "option_id_is_required"}
			}
			if strings.TrimSpace(opt.OptionName) == "" {
				return &httpErrors.BadRequestError{Message: "option_name_is_required"}
			}
			if opt.OptionPrice < 0 {
				return &httpErrors.BadRequestError{Message: "option_price_cannot_be_negative"}
			}
		}
	}

	return nil
}

// ToModel converts the HTTP request to a domain model.
func (r *CreateOrderRequest) ToModel() *models.Order {
	return r.Order.ToModel()
}

// ToModel converts the order to a domain model.
// Maps all snapshot fields for historical preservation.
func (r *OrderRequest) ToModel() *models.Order {
	order := &models.Order{
		Customer: &models.Customer{
			Name:  r.Customer.Name,
			Phone: r.Customer.Phone,
			Email: r.Customer.Email,
		},
		PaymentMethod: &models.PaymentMethod{
			ID:   r.PaymentMethod.ID,
			Name: r.PaymentMethod.Name,
			Code: models.PaymentMethodCode(r.PaymentMethod.Code),
		},
		DeliveryMethod: &models.DeliveryMethod{
			ID:   r.DeliveryMethod.ID,
			Name: r.DeliveryMethod.Name,
			Code: models.DeliveryMethodCode(r.DeliveryMethod.Code),
		},
		ShippingCost: r.DeliveryMethod.ShippingCost,
		Subtotal:     r.Subtotal,
		Total:        r.Total,
		Items:        make([]*models.OrderItem, 0, len(r.Items)),
	}

	// Set delivery zone if provided
	if r.DeliveryMethod.DeliveryZone != nil {
		order.DeliveryMethod.DeliveryZones = []*models.DeliveryZone{
			{
				ID:    r.DeliveryMethod.DeliveryZone.ID,
				Name:  r.DeliveryMethod.DeliveryZone.Name,
				Price: r.DeliveryMethod.DeliveryZone.Price,
			},
		}
	}

	// Set customer address if provided
	if r.Customer.Address != nil {
		order.Customer.Address = &models.Address{
			Name:    r.Customer.Address.Name,
			PlaceID: r.Customer.Address.PlaceID,
			Lat:     r.Customer.Address.Lat,
			Lng:     r.Customer.Address.Lng,
		}
	}

	// Convert items with snapshot data
	for _, item := range r.Items {
		// Build product with snapshot and image
		product := &models.Product{
			ID:               item.Product.ID,
			Name:             item.Product.Name,
			Price:            item.Product.Price,
			IsPromotional:    item.Product.IsPromotional,
			PromotionalPrice: item.Product.PromotionalPrice,
			Variants:         make([]*models.Variant, 0),
		}

		// Add image if provided
		if item.Product.ImageURL != "" {
			product.Images = []*models.Image{{URL: item.Product.ImageURL}}
		}

		orderItem := &models.OrderItem{
			Product:   product,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}

		// Group selected options by variant with snapshot data
		variantMap := make(map[int]*models.Variant)
		for _, opt := range item.SelectedOptions {
			variant, exists := variantMap[opt.VariantID]
			if !exists {
				variant = &models.Variant{
					ID:      opt.VariantID,
					Name:    opt.VariantName,
					Options: make([]*models.Option, 0),
				}
				variantMap[opt.VariantID] = variant
			}
			variant.Options = append(variant.Options, &models.Option{
				ID:    opt.OptionID,
				Name:  opt.OptionName,
				Price: opt.OptionPrice,
			})
		}

		// Convert map to slice
		for _, variant := range variantMap {
			orderItem.Product.Variants = append(orderItem.Product.Variants, variant)
		}

		order.Items = append(order.Items, orderItem)
	}

	return order
}
