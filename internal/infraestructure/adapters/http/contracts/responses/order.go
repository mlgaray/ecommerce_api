package responses

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// OrderResponse represents an order in HTTP responses.
// Used by all order endpoints (Create, Update, List, GetByID).
// List endpoints populate partial data (no items, minimal nested objects).
// Detail endpoints populate all fields.
type OrderResponse struct {
	ID             int                          `json:"id"`
	OrderNumber    string                       `json:"order_number"`
	Status         string                       `json:"status"`
	Customer       OrderCustomerResponse        `json:"customer"`
	PaymentMethod  *OrderPaymentMethodResponse  `json:"payment_method,omitempty"`
	DeliveryMethod *OrderDeliveryMethodResponse `json:"delivery_method,omitempty"`
	Items          []*OrderItemResponse         `json:"items"`
	ItemsCount     int                          `json:"items_count"`
	Subtotal       float64                      `json:"subtotal"`
	ShippingCost   float64                      `json:"shipping_cost"`
	Total          float64                      `json:"total"`
	CreatedAt      time.Time                    `json:"created_at"`
}

// OrderCustomerResponse represents customer data in order responses.
type OrderCustomerResponse struct {
	Name    string                `json:"name"`
	Phone   string                `json:"phone,omitempty"`
	Email   string                `json:"email,omitempty"`
	Address *OrderAddressResponse `json:"address,omitempty"`
}

// OrderAddressResponse represents address data in order responses.
type OrderAddressResponse struct {
	Name    string  `json:"name"`
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}

// OrderPaymentMethodResponse represents payment method data in order responses.
type OrderPaymentMethodResponse struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

// OrderDeliveryMethodResponse represents delivery method data in order responses.
type OrderDeliveryMethodResponse struct {
	ID           int                        `json:"id,omitempty"`
	Name         string                     `json:"name"`
	Code         string                     `json:"code,omitempty"`
	DeliveryZone *OrderDeliveryZoneResponse `json:"delivery_zone,omitempty"`
}

// OrderDeliveryZoneResponse represents a delivery zone in order responses.
type OrderDeliveryZoneResponse struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// OrderItemResponse represents an item in order responses.
type OrderItemResponse struct {
	Product         OrderProductResponse           `json:"product"`
	Quantity        int                            `json:"quantity"`
	UnitPrice       float64                        `json:"unit_price"`
	TotalPrice      float64                        `json:"total_price"`
	SelectedOptions []*OrderSelectedOptionResponse `json:"selected_options"`
}

// OrderProductResponse represents product snapshot data in order item responses.
type OrderProductResponse struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	ImageURL         string  `json:"image_url,omitempty"`
	Price            float64 `json:"price"`
	IsPromotional    bool    `json:"is_promotional"`
	PromotionalPrice float64 `json:"promotional_price,omitempty"`
}

// OrderSelectedOptionResponse represents a selected option in order item responses.
type OrderSelectedOptionResponse struct {
	VariantID   int     `json:"variant_id"`
	VariantName string  `json:"variant_name"`
	OptionID    int     `json:"option_id"`
	OptionName  string  `json:"option_name"`
	OptionPrice float64 `json:"option_price"`
}

// OrderResponseFromModel converts a domain Order to an OrderResponse.
//
//nolint:gocyclo // Complexity is intentional for readability - linear field mappings for nested structs (customer, payment, delivery, items with flattened options)
func OrderResponseFromModel(order *models.Order) *OrderResponse {
	if order == nil {
		return nil
	}

	response := &OrderResponse{
		ID:           order.ID,
		OrderNumber:  order.OrderNumber,
		Status:       string(order.Status),
		ItemsCount:   order.ItemsCount,
		Subtotal:     order.Subtotal,
		ShippingCost: order.ShippingCost,
		Total:        order.Total,
		CreatedAt:    order.CreatedAt,
	}

	// Convert customer
	if order.Customer != nil {
		response.Customer = OrderCustomerResponse{
			Name:  order.Customer.Name,
			Phone: order.Customer.Phone,
			Email: order.Customer.Email,
		}
		if order.Customer.Address != nil {
			response.Customer.Address = &OrderAddressResponse{
				Name:    order.Customer.Address.Name,
				PlaceID: order.Customer.Address.PlaceID,
				Lat:     order.Customer.Address.Lat,
				Lng:     order.Customer.Address.Lng,
			}
		}
	}

	// Convert payment method
	if order.PaymentMethod != nil {
		response.PaymentMethod = &OrderPaymentMethodResponse{
			ID:   order.PaymentMethod.ID,
			Name: order.PaymentMethod.Name,
			Code: string(order.PaymentMethod.Code),
		}
	}

	// Convert delivery method
	if order.DeliveryMethod != nil {
		dm := &OrderDeliveryMethodResponse{
			ID:   order.DeliveryMethod.ID,
			Name: order.DeliveryMethod.Name,
			Code: string(order.DeliveryMethod.Code),
		}
		if len(order.DeliveryMethod.DeliveryZones) > 0 && order.DeliveryMethod.DeliveryZones[0] != nil {
			dm.DeliveryZone = &OrderDeliveryZoneResponse{
				ID:    order.DeliveryMethod.DeliveryZones[0].ID,
				Name:  order.DeliveryMethod.DeliveryZones[0].Name,
				Price: order.DeliveryMethod.DeliveryZones[0].Price,
			}
		}
		response.DeliveryMethod = dm
	}

	// Convert items with flattened selected options
	response.Items = make([]*OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		itemResponse := &OrderItemResponse{
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		}

		// Convert product snapshot
		if item.Product != nil {
			itemResponse.Product = OrderProductResponse{
				ID:               item.Product.ID,
				Name:             item.Product.Name,
				Price:            item.Product.Price,
				IsPromotional:    item.Product.IsPromotional,
				PromotionalPrice: item.Product.PromotionalPrice,
			}
			if len(item.Product.Images) > 0 {
				itemResponse.Product.ImageURL = item.Product.Images[0].URL
			}
		}

		// Flatten variants/options back to selected_options format
		itemResponse.SelectedOptions = make([]*OrderSelectedOptionResponse, 0)
		if item.Product != nil {
			for _, variant := range item.Product.Variants {
				for _, option := range variant.Options {
					itemResponse.SelectedOptions = append(itemResponse.SelectedOptions, &OrderSelectedOptionResponse{
						VariantID:   variant.ID,
						VariantName: variant.Name,
						OptionID:    option.ID,
						OptionName:  option.Name,
						OptionPrice: option.Price,
					})
				}
			}
		}

		response.Items = append(response.Items, itemResponse)
	}

	return response
}

// OrderResponsesFromModels converts a slice of domain Orders to OrderResponses.
func OrderResponsesFromModels(orders []*models.Order) []*OrderResponse {
	if orders == nil {
		return nil
	}

	responses := make([]*OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = OrderResponseFromModel(o)
	}
	return responses
}

// CreateOrderResponse represents the response for order creation.
type CreateOrderResponse struct {
	Order *OrderResponse `json:"order"`
}

// ListOrdersResponse represents the HTTP response for list orders.
type ListOrdersResponse struct {
	Orders     []*OrderResponse `json:"orders"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
	TotalCount *int             `json:"total_count,omitempty"`
}

// NewListOrdersResponse creates a ListOrdersResponse from domain models.
func NewListOrdersResponse(
	orders []*models.Order,
	nextCursor string,
	hasMore bool,
	totalCount *int,
) *ListOrdersResponse {
	return &ListOrdersResponse{
		Orders:     OrderResponsesFromModels(orders),
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}
}
