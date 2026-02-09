package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidCreateOrderRequest() *CreateOrderRequest {
	return &CreateOrderRequest{
		Order: OrderRequest{
			Customer: OrderCustomerRequest{
				Name:  "John Doe",
				Phone: "123456789",
				Email: "john@example.com",
			},
			PaymentMethod: OrderPaymentMethodRequest{
				ID:   1,
				Name: "Efectivo",
				Code: "cash",
			},
			DeliveryMethod: OrderDeliveryMethodRequest{
				ID:           1,
				Name:         "Retiro en local",
				Code:         "pickup",
				ShippingCost: 0,
			},
			Items: []OrderItemRequest{
				{
					Product: OrderProductRequest{
						ID:    1,
						Name:  "Pizza Muzzarella",
						Price: 1500.00,
					},
					Quantity:  2,
					UnitPrice: 1500.00,
				},
			},
			Subtotal: 3000.00,
			Total:    3000.00,
		},
	}
}

// =============================================================================
// Validate Tests - Customer
// =============================================================================

func TestCreateOrderRequest_Validate_Customer(t *testing.T) {
	t.Run("when customer name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Customer.Name = ""

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "customer_name_is_required", badRequestErr.Message)
	})

	t.Run("when customer name is whitespace then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Customer.Name = "   "

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "customer_name_is_required", badRequestErr.Message)
	})

	t.Run("when customer name is valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Customer.Name = "John Doe"

		err := request.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Payment Method
// =============================================================================

func TestCreateOrderRequest_Validate_PaymentMethod(t *testing.T) {
	t.Run("when payment method ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.PaymentMethod.ID = 0

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "payment_method_id_is_required", badRequestErr.Message)
	})

	t.Run("when payment method ID is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.PaymentMethod.ID = -1

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "payment_method_id_is_required", badRequestErr.Message)
	})

	t.Run("when payment method name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.PaymentMethod.Name = ""

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "payment_method_name_is_required", badRequestErr.Message)
	})

	t.Run("when payment method code is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.PaymentMethod.Code = ""

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "payment_method_code_is_required", badRequestErr.Message)
	})

	t.Run("when payment method is valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()

		err := request.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Delivery Method
// =============================================================================

func TestCreateOrderRequest_Validate_DeliveryMethod(t *testing.T) {
	t.Run("when delivery method ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.ID = 0

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_method_id_is_required", badRequestErr.Message)
	})

	t.Run("when delivery method ID is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.ID = -1

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_method_id_is_required", badRequestErr.Message)
	})

	t.Run("when delivery method name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.Name = ""

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_method_name_is_required", badRequestErr.Message)
	})

	t.Run("when delivery method code is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.Code = ""

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_method_code_is_required", badRequestErr.Message)
	})

	t.Run("when delivery method is valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()

		err := request.Validate()

		assert.NoError(t, err)
	})

	t.Run("when shipping cost is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.ShippingCost = -10.00

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "shipping_cost_cannot_be_negative", badRequestErr.Message)
	})

	t.Run("when shipping cost is zero then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.ShippingCost = 0

		err := request.Validate()

		assert.NoError(t, err)
	})

	t.Run("when shipping cost is positive then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.ShippingCost = 100.00

		err := request.Validate()

		assert.NoError(t, err)
	})

	t.Run("when delivery zone ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &OrderDeliveryZoneRequest{ID: 0, Name: "Zona Norte", Price: 500}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_zone_id_is_required", badRequestErr.Message)
	})

	t.Run("when delivery zone name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &OrderDeliveryZoneRequest{ID: 1, Name: "", Price: 500}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "delivery_zone_name_is_required", badRequestErr.Message)
	})

	t.Run("when delivery zone is valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &OrderDeliveryZoneRequest{ID: 1, Name: "Zona Norte", Price: 500}

		err := request.Validate()

		assert.NoError(t, err)
	})

	t.Run("when delivery zone is nil then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = nil

		err := request.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Items
// =============================================================================

func TestCreateOrderRequest_Validate_Items(t *testing.T) {
	t.Run("when items is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "order_must_have_at_least_one_item", badRequestErr.Message)
	})

	t.Run("when items is nil then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = nil

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "order_must_have_at_least_one_item", badRequestErr.Message)
	})

	t.Run("when item product ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 0, Name: "Pizza", Price: 1500}, Quantity: 1, UnitPrice: 1500},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "product_id_is_required", badRequestErr.Message)
	})

	t.Run("when item product name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "", Price: 1500}, Quantity: 1, UnitPrice: 1500},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "product_name_is_required", badRequestErr.Message)
	})

	t.Run("when item product base price is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "Pizza", Price: -100}, Quantity: 1, UnitPrice: 1500},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "product_price_cannot_be_negative", badRequestErr.Message)
	})

	t.Run("when item quantity is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500}, Quantity: 0, UnitPrice: 1500},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "quantity_must_be_positive", badRequestErr.Message)
	})

	t.Run("when item quantity is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500}, Quantity: -1, UnitPrice: 1500},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "quantity_must_be_positive", badRequestErr.Message)
	})

	t.Run("when unit price is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500}, Quantity: 1, UnitPrice: -100},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "unit_price_cannot_be_negative", badRequestErr.Message)
	})

	t.Run("when multiple items are valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{Product: OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500}, Quantity: 2, UnitPrice: 1500},
			{Product: OrderProductRequest{ID: 2, Name: "Empanadas", Price: 300}, Quantity: 3, UnitPrice: 300},
			{Product: OrderProductRequest{ID: 3, Name: "Bebida", Price: 500}, Quantity: 1, UnitPrice: 500},
		}

		err := request.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Selected Options
// =============================================================================

func TestCreateOrderRequest_Validate_SelectedOptions(t *testing.T) {
	t.Run("when selected option variant ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 0, VariantName: "Tamaño", OptionID: 1, OptionName: "Grande", OptionPrice: 300},
				},
			},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "variant_id_is_required", badRequestErr.Message)
	})

	t.Run("when selected option variant name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "", OptionID: 1, OptionName: "Grande", OptionPrice: 300},
				},
			},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "variant_name_is_required", badRequestErr.Message)
	})

	t.Run("when selected option option ID is zero then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 0, OptionName: "Grande", OptionPrice: 300},
				},
			},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "option_id_is_required", badRequestErr.Message)
	})

	t.Run("when selected option option name is empty then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 1, OptionName: "", OptionPrice: 300},
				},
			},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "option_name_is_required", badRequestErr.Message)
	})

	t.Run("when selected option option price is negative then returns bad request error", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 1, OptionName: "Grande", OptionPrice: -100},
				},
			},
		}

		err := request.Validate()

		assert.Error(t, err)
		badRequestErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "option_price_cannot_be_negative", badRequestErr.Message)
	})

	t.Run("when selected options are valid then validation passes", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 10, OptionName: "Grande", OptionPrice: 300},
					{VariantID: 2, VariantName: "Masa", OptionID: 20, OptionName: "Crocante", OptionPrice: 0},
				},
			},
		}

		err := request.Validate()

		assert.NoError(t, err)
	})
}

// =============================================================================
// ToModel Tests
// =============================================================================

func TestCreateOrderRequest_ToModel(t *testing.T) {
	t.Run("converts basic fields correctly", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Customer.Name = "John Doe"
		request.Order.Customer.Phone = "123456789"
		request.Order.Customer.Email = "john@example.com"
		request.Order.PaymentMethod = OrderPaymentMethodRequest{ID: 1, Name: "Efectivo", Code: "cash"}
		request.Order.DeliveryMethod = OrderDeliveryMethodRequest{ID: 2, Name: "Envío", Code: "delivery", ShippingCost: 10.00}

		order := request.ToModel()

		assert.NotNil(t, order.Customer)
		assert.Equal(t, "John Doe", order.Customer.Name)
		assert.Equal(t, "123456789", order.Customer.Phone)
		assert.Equal(t, "john@example.com", order.Customer.Email)
		assert.NotNil(t, order.PaymentMethod)
		assert.Equal(t, 1, order.PaymentMethod.ID)
		assert.Equal(t, "Efectivo", order.PaymentMethod.Name)
		assert.Equal(t, "cash", string(order.PaymentMethod.Code))
		assert.NotNil(t, order.DeliveryMethod)
		assert.Equal(t, 2, order.DeliveryMethod.ID)
		assert.Equal(t, "Envío", order.DeliveryMethod.Name)
		assert.Equal(t, "delivery", string(order.DeliveryMethod.Code))
		assert.Equal(t, 10.00, order.ShippingCost)
	})

	t.Run("converts delivery zone correctly when provided", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &OrderDeliveryZoneRequest{ID: 5, Name: "Zona Norte", Price: 500}

		order := request.ToModel()

		assert.NotNil(t, order.DeliveryMethod.DeliveryZones)
		assert.Len(t, order.DeliveryMethod.DeliveryZones, 1)
		assert.Equal(t, 5, order.DeliveryMethod.DeliveryZones[0].ID)
		assert.Equal(t, "Zona Norte", order.DeliveryMethod.DeliveryZones[0].Name)
		assert.Equal(t, 500.0, order.DeliveryMethod.DeliveryZones[0].Price)
	})

	t.Run("converts address correctly when provided", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Customer.Address = &OrderAddressRequest{
			Name:    "123 Main St",
			PlaceID: "place123",
			Lat:     -34.603722,
			Lng:     -58.381592,
		}

		order := request.ToModel()

		assert.NotNil(t, order.Customer.Address)
		assert.Equal(t, "123 Main St", order.Customer.Address.Name)
		assert.Equal(t, "place123", order.Customer.Address.PlaceID)
		assert.Equal(t, -34.603722, order.Customer.Address.Lat)
		assert.Equal(t, -58.381592, order.Customer.Address.Lng)
	})

	t.Run("converts items with product snapshot correctly", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", ImageURL: "http://img.com/pizza.jpg", Price: 1500, IsPromotional: true, PromotionalPrice: 1200},
				Quantity:  2,
				UnitPrice: 1200,
			},
			{
				Product:   OrderProductRequest{ID: 3, Name: "Empanadas", Price: 300},
				Quantity:  5,
				UnitPrice: 300,
			},
		}

		order := request.ToModel()

		assert.Len(t, order.Items, 2)
		// First item with promotion and image
		assert.Equal(t, 1, order.Items[0].Product.ID)
		assert.Equal(t, "Pizza", order.Items[0].Product.Name)
		assert.Equal(t, 1500.0, order.Items[0].Product.Price)
		assert.True(t, order.Items[0].Product.IsPromotional)
		assert.Equal(t, 1200.0, order.Items[0].Product.PromotionalPrice)
		assert.Len(t, order.Items[0].Product.Images, 1)
		assert.Equal(t, "http://img.com/pizza.jpg", order.Items[0].Product.Images[0].URL)
		assert.Equal(t, 2, order.Items[0].Quantity)
		assert.Equal(t, 1200.0, order.Items[0].UnitPrice)
		// Second item without promotion
		assert.Equal(t, 3, order.Items[1].Product.ID)
		assert.Equal(t, "Empanadas", order.Items[1].Product.Name)
		assert.Equal(t, 300.0, order.Items[1].Product.Price)
		assert.False(t, order.Items[1].Product.IsPromotional)
		assert.Equal(t, 5, order.Items[1].Quantity)
		assert.Equal(t, 300.0, order.Items[1].UnitPrice)
	})

	t.Run("converts subtotal and total correctly", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Subtotal = 200.00
		request.Order.Total = 220.00

		order := request.ToModel()

		assert.Equal(t, 200.00, order.Subtotal)
		assert.Equal(t, 220.00, order.Total)
	})

	t.Run("converts selected options grouped by variant with snapshot", func(t *testing.T) {
		request := newValidCreateOrderRequest()
		request.Order.Items = []OrderItemRequest{
			{
				Product:   OrderProductRequest{ID: 1, Name: "Pizza", Price: 1500},
				Quantity:  1,
				UnitPrice: 1800,
				SelectedOptions: []OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 10, OptionName: "Grande", OptionPrice: 300},
					{VariantID: 1, VariantName: "Tamaño", OptionID: 11, OptionName: "Familiar", OptionPrice: 500},
					{VariantID: 2, VariantName: "Masa", OptionID: 20, OptionName: "Crocante", OptionPrice: 0},
				},
			},
		}

		order := request.ToModel()

		assert.Len(t, order.Items, 1)
		// Two variants (1 and 2)
		assert.Len(t, order.Items[0].Product.Variants, 2)

		// Find variant with ID 1
		var variantTamanio *models.Variant
		var variantMasa *models.Variant
		for _, v := range order.Items[0].Product.Variants {
			if v.ID == 1 {
				variantTamanio = v
			} else if v.ID == 2 {
				variantMasa = v
			}
		}

		// Variant "Tamaño" has 2 options
		assert.NotNil(t, variantTamanio)
		assert.Equal(t, "Tamaño", variantTamanio.Name)
		assert.Len(t, variantTamanio.Options, 2)

		// Variant "Masa" has 1 option
		assert.NotNil(t, variantMasa)
		assert.Equal(t, "Masa", variantMasa.Name)
		assert.Len(t, variantMasa.Options, 1)
		assert.Equal(t, "Crocante", variantMasa.Options[0].Name)
		assert.Equal(t, 0.0, variantMasa.Options[0].Price)
	})
}

// =============================================================================
// Full Validation Flow Tests
// =============================================================================

func TestCreateOrderRequest_Validate_FullFlow(t *testing.T) {
	t.Run("when all fields are valid then validation passes", func(t *testing.T) {
		request := &CreateOrderRequest{
			Order: OrderRequest{
				Customer: OrderCustomerRequest{
					Name:  "Juan Pérez",
					Phone: "+5491123456789",
					Email: "juan@test.com",
					Address: &OrderAddressRequest{
						Name:    "Av. Corrientes 1234, CABA",
						PlaceID: "ChIJ...",
						Lat:     -34.603722,
						Lng:     -58.381592,
					},
				},
				PaymentMethod: OrderPaymentMethodRequest{
					ID:   1,
					Name: "Efectivo",
					Code: "cash",
				},
				DeliveryMethod: OrderDeliveryMethodRequest{
					ID:           2,
					Name:         "Envío a domicilio",
					Code:         "delivery",
					ShippingCost: 500.0,
					DeliveryZone: &OrderDeliveryZoneRequest{
						ID:    3,
						Name:  "Zona Norte",
						Price: 500.0,
					},
				},
				Items: []OrderItemRequest{
					{
						Product: OrderProductRequest{
							ID:    1,
							Name:  "Pizza Muzzarella",
							Price: 1500.0,
						},
						Quantity:  2,
						UnitPrice: 1800.0,
						SelectedOptions: []OrderSelectedOptionRequest{
							{
								VariantID:   1,
								VariantName: "Tamaño",
								OptionID:    10,
								OptionName:  "Grande",
								OptionPrice: 300.0,
							},
						},
					},
					{
						Product: OrderProductRequest{
							ID:    2,
							Name:  "Empanadas x6",
							Price: 1200.0,
						},
						Quantity:  1,
						UnitPrice: 1200.0,
					},
				},
				Subtotal: 4800.0,
				Total:    5300.0,
			},
		}

		err := request.Validate()

		assert.NoError(t, err)
	})
}
