package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidOrderRequest() *contracts.CreateOrderRequest {
	return &contracts.CreateOrderRequest{
		Order: contracts.OrderRequest{
			Customer: contracts.OrderCustomerRequest{
				Name:  "John Doe",
				Phone: "123456789",
				Email: "john@example.com",
			},
			PaymentMethod: contracts.OrderPaymentMethodRequest{
				ID:   1,
				Name: "Efectivo",
				Code: "cash",
			},
			DeliveryMethod: contracts.OrderDeliveryMethodRequest{
				ID:           1,
				Name:         "Retiro en local",
				Code:         "pickup",
				ShippingCost: 10.00,
			},
			Items: []contracts.OrderItemRequest{
				{
					Product: contracts.OrderProductRequest{
						ID:    1,
						Name:  "Test Product",
						Price: 50.00,
					},
					Quantity:  2,
					UnitPrice: 50.00,
				},
			},
			Subtotal: 100.00,
			Total:    110.00,
		},
	}
}

func newCreatedOrder() *models.Order {
	return &models.Order{
		ID:          1,
		OrderNumber: "ORD-2024-00001",
		Status:      models.OrderStatusPending,
		Store: &models.Store{
			ID:   1,
			Name: "Test Store",
			Slug: "test-store",
		},
		Customer: &models.Customer{
			Name:  "John Doe",
			Phone: "123456789",
			Email: "john@example.com",
		},
		Items: []*models.OrderItem{
			{
				ID:         1,
				Product:    &models.Product{ID: 1, Name: "Test Product"},
				Quantity:   2,
				UnitPrice:  50.00,
				TotalPrice: 100.00,
			},
		},
		Subtotal:     100.00,
		ShippingCost: 10.00,
		Total:        110.00,
	}
}

// =============================================================================
// Create Tests - Success
// =============================================================================

func TestOrderHandler_Create(t *testing.T) {
	t.Run("when request is valid then returns 201 with created order", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		createdOrder := newCreatedOrder()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response models.Order
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, createdOrder.ID, response.ID)
		assert.Equal(t, createdOrder.OrderNumber, response.OrderNumber)
		assert.Equal(t, createdOrder.Total, response.Total)
	})
}

// =============================================================================
// Create Tests - Invalid Slug
// =============================================================================

func TestOrderHandler_Create_InvalidSlug(t *testing.T) {
	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/stores//orders", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": ""})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when slug is whitespace then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/stores/whitespace/orders", nil)
		req = mux.SetURLVars(req, map[string]string{"slug": "   "})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Create Tests - Invalid JSON
// =============================================================================

func TestOrderHandler_Create_InvalidJSON(t *testing.T) {
	t.Run("when body is not valid JSON then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when body is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", nil)
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Create Tests - Request Validation
// =============================================================================

func TestOrderHandler_Create_RequestValidation(t *testing.T) {
	t.Run("when customer name is missing then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.Customer.Name = ""

		handler := NewOrderHandler(nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when payment method ID is missing then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.PaymentMethod.ID = 0

		handler := NewOrderHandler(nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when items is empty then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.Items = []contracts.OrderItemRequest{}

		handler := NewOrderHandler(nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Create Tests - Store Not Found
// =============================================================================

func TestOrderHandler_Create_StoreNotFound(t *testing.T) {
	t.Run("when store not found then returns 404", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "non-existent-store").
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.StoreNotFound})

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/non-existent-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "non-existent-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// =============================================================================
// Create Tests - Validation Error
// =============================================================================

func TestOrderHandler_Create_ValidationError(t *testing.T) {
	t.Run("when payment method not found then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(nil, &domainErrors.ValidationError{Message: domainErrors.PaymentMethodNotFound})

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when insufficient stock then returns 422", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(nil, &domainErrors.BusinessRuleError{Message: domainErrors.InsufficientStock})

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("when shipping cost mismatch then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(nil, &domainErrors.ValidationError{Message: domainErrors.ShippingCostMismatch})

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Create Tests - With Selected Options
// =============================================================================

func TestOrderHandler_Create_WithSelectedOptions(t *testing.T) {
	t.Run("when request has selected options then passes to use case", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.Items = []contracts.OrderItemRequest{
			{
				Product: contracts.OrderProductRequest{
					ID:    1,
					Name:  "Pizza",
					Price: 50.00,
				},
				Quantity:  1,
				UnitPrice: 80.00,
				SelectedOptions: []contracts.OrderSelectedOptionRequest{
					{VariantID: 1, VariantName: "Tamaño", OptionID: 10, OptionName: "Grande", OptionPrice: 20.00},
					{VariantID: 2, VariantName: "Masa", OptionID: 20, OptionName: "Crocante", OptionPrice: 10.00},
				},
			},
		}
		request.Order.Subtotal = 80.00
		request.Order.Total = 90.00
		createdOrder := newCreatedOrder()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

// =============================================================================
// Create Tests - With Customer Address
// =============================================================================

func TestOrderHandler_Create_WithCustomerAddress(t *testing.T) {
	t.Run("when request has customer address then passes to use case", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.Customer.Address = &contracts.OrderAddressRequest{
			Name:    "123 Main St",
			PlaceID: "place123",
			Lat:     -34.603722,
			Lng:     -58.381592,
		}
		createdOrder := newCreatedOrder()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

// =============================================================================
// Create Tests - With Delivery Zone
// =============================================================================

func TestOrderHandler_Create_WithDeliveryZone(t *testing.T) {
	t.Run("when request has delivery zone then passes to use case", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &contracts.OrderDeliveryZoneRequest{
			ID:    5,
			Name:  "Zona Norte",
			Price: 25.00,
		}
		request.Order.DeliveryMethod.ShippingCost = 25.00
		createdOrder := newCreatedOrder()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.AnythingOfType("*models.Order"), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}
