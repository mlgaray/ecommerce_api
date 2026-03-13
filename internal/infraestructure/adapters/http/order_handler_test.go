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
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/requests"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidOrderRequest() *requests.CreateOrderRequest {
	return &requests.CreateOrderRequest{
		Order: requests.OrderRequest{
			Customer: requests.OrderCustomerRequest{
				Name:  "John Doe",
				Phone: "123456789",
				Email: "john@example.com",
			},
			PaymentMethod: requests.OrderPaymentMethodRequest{
				ID:   1,
				Name: "Efectivo",
				Code: "cash",
			},
			DeliveryMethod: requests.OrderDeliveryMethodRequest{
				ID:           1,
				Name:         "Retiro en local",
				Code:         "pickup",
				ShippingCost: 10.00,
			},
			Items: []requests.OrderItemRequest{
				{
					Product: requests.OrderProductRequest{
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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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

		var response responses.CreateOrderResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Order) {
			assert.Equal(t, createdOrder.ID, response.Order.ID)
			assert.Equal(t, createdOrder.OrderNumber, response.Order.OrderNumber)
			assert.Equal(t, createdOrder.Total, response.Order.Total)
		}
	})
}

// =============================================================================
// Create Tests - Invalid Slug
// =============================================================================

func TestOrderHandler_Create_InvalidSlug(t *testing.T) {
	t.Run("when slug is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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
		request.Order.Items = []requests.OrderItemRequest{}

		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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
		request.Order.Items = []requests.OrderItemRequest{
			{
				Product: requests.OrderProductRequest{
					ID:    1,
					Name:  "Pizza",
					Price: 50.00,
				},
				Quantity:  1,
				UnitPrice: 80.00,
				SelectedOptions: []requests.OrderSelectedOptionRequest{
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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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
		request.Order.Customer.Address = &requests.OrderAddressRequest{
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

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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
	t.Run("when request has delivery zone then maps zone data to domain model", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		request.Order.DeliveryMethod.DeliveryZone = &requests.OrderDeliveryZoneRequest{
			ID:    5,
			Name:  "Zona Norte",
			Price: 25.00,
		}
		request.Order.DeliveryMethod.ShippingCost = 25.00
		createdOrder := newCreatedOrder()
		createdOrder.DeliveryMethod = &models.DeliveryMethod{
			ID:   1,
			Name: "Retiro en local",
			Code: "pickup",
			DeliveryZones: []*models.DeliveryZone{
				{ID: 5, Name: "Zona Norte", Price: 25.00},
			},
		}

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.MatchedBy(func(order *models.Order) bool {
				return order.DeliveryMethod != nil &&
					len(order.DeliveryMethod.DeliveryZones) == 1 &&
					order.DeliveryMethod.DeliveryZones[0].ID == 5 &&
					order.DeliveryMethod.DeliveryZones[0].Name == "Zona Norte" &&
					order.DeliveryMethod.DeliveryZones[0].Price == 25.00
			}), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPost, "/stores/test-store/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"slug": "test-store"})
		rr := httptest.NewRecorder()

		// Act
		handler.Create(rr, req)

		// Assert
		assert.Equal(t, http.StatusCreated, rr.Code)

		var response responses.CreateOrderResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Order) && assert.NotNil(t, response.Order.DeliveryMethod) {
			assert.NotNil(t, response.Order.DeliveryMethod.DeliveryZone)
			assert.Equal(t, 5, response.Order.DeliveryMethod.DeliveryZone.ID)
			assert.Equal(t, "Zona Norte", response.Order.DeliveryMethod.DeliveryZone.Name)
			assert.Equal(t, 25.00, response.Order.DeliveryMethod.DeliveryZone.Price)
		}
	})

	t.Run("when request has no delivery zone then zone is nil in domain model", func(t *testing.T) {
		// Arrange
		request := newValidOrderRequest()
		// No DeliveryZone set (default is nil)
		createdOrder := newCreatedOrder()

		useCaseMock := mocks.NewCreateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, mock.MatchedBy(func(order *models.Order) bool {
				return order.DeliveryMethod != nil &&
					len(order.DeliveryMethod.DeliveryZones) == 0
			}), "test-store").
			Return(createdOrder, nil)

		handler := NewOrderHandler(useCaseMock, nil, nil, nil, nil, nil)

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
// UpdateStatus Test Helpers
// =============================================================================

func newUpdateStatusRequest(status string) map[string]string {
	return map[string]string{"status": status}
}

func newConfirmedOrder() *models.Order {
	order := newCreatedOrder()
	order.Status = models.OrderStatusConfirmed
	return order
}

// =============================================================================
// UpdateStatus Tests - Success
// =============================================================================

func TestOrderHandler_UpdateStatus(t *testing.T) {
	t.Run("when valid transition then returns 200 with updated order", func(t *testing.T) {
		// Arrange
		updatedOrder := newConfirmedOrder()

		useCaseMock := mocks.NewUpdateOrderStatusUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, models.OrderStatusConfirmed).
			Return(updatedOrder, nil)

		handler := NewOrderHandler(nil, nil, nil, useCaseMock, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest("confirmed"))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response responses.UpdateOrderStatusResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Order) {
			assert.Equal(t, "confirmed", response.Order.Status)
		}
	})
}

// =============================================================================
// UpdateStatus Tests - Invalid Request
// =============================================================================

func TestOrderHandler_UpdateStatus_InvalidRequest(t *testing.T) {
	t.Run("when status is empty then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest(""))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when status is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest("invalid_status"))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when body is invalid JSON then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// UpdateStatus Tests - Domain Errors
// =============================================================================

func TestOrderHandler_UpdateStatus_DomainErrors(t *testing.T) {
	t.Run("when order not found then returns 404", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewUpdateOrderStatusUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 999, models.OrderStatusConfirmed).
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.OrderNotFound})

		handler := NewOrderHandler(nil, nil, nil, useCaseMock, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest("confirmed"))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/999/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when invalid transition then returns 400", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewUpdateOrderStatusUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, models.OrderStatusCompleted).
			Return(nil, &domainErrors.ValidationError{Message: domainErrors.InvalidOrderTransition})

		handler := NewOrderHandler(nil, nil, nil, useCaseMock, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest("completed"))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when concurrent modification then returns 409", func(t *testing.T) {
		// Arrange
		useCaseMock := mocks.NewUpdateOrderStatusUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, models.OrderStatusConfirmed).
			Return(nil, &domainErrors.ConcurrentModificationError{Message: domainErrors.OrderConcurrentModification})

		handler := NewOrderHandler(nil, nil, nil, useCaseMock, nil, nil)

		body, _ := json.Marshal(newUpdateStatusRequest("confirmed"))
		req := httptest.NewRequest(http.MethodPatch, "/shops/1/orders/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.UpdateStatus(rr, req)

		// Assert
		assert.Equal(t, http.StatusConflict, rr.Code)
	})
}

// =============================================================================
// Update Test Helpers
// =============================================================================

func newValidUpdateOrderRequest() *requests.UpdateOrderRequest {
	return &requests.UpdateOrderRequest{
		Order: requests.UpdateOrderDataRequest{
			Customer: requests.OrderCustomerRequest{
				Name:  "Updated Customer",
				Phone: "987654321",
				Email: "updated@example.com",
			},
			PaymentMethod: &requests.OrderPaymentMethodRequest{
				ID:   2,
				Name: "Transferencia",
				Code: "transfer",
			},
			DeliveryMethod: &requests.OrderDeliveryMethodRequest{
				ID:           1,
				Name:         "Delivery",
				Code:         "delivery",
				ShippingCost: 15.00,
			},
			Items: []requests.OrderItemRequest{
				{
					Product: requests.OrderProductRequest{
						ID:    1,
						Name:  "Updated Product",
						Price: 75.00,
					},
					Quantity:  3,
					UnitPrice: 75.00,
				},
			},
			Subtotal: 225.00,
			Total:    240.00,
		},
	}
}

func newUpdatedOrder() *models.Order {
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
			Name:  "Updated Customer",
			Phone: "987654321",
			Email: "updated@example.com",
		},
		Items: []*models.OrderItem{
			{
				ID:         1,
				Product:    &models.Product{ID: 1, Name: "Updated Product"},
				Quantity:   3,
				UnitPrice:  75.00,
				TotalPrice: 225.00,
			},
		},
		Subtotal:     225.00,
		ShippingCost: 15.00,
		Total:        240.00,
	}
}

// =============================================================================
// Update Tests - Success
// =============================================================================

func TestOrderHandler_Update(t *testing.T) {
	t.Run("when request is valid then returns 200 with updated order", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()
		updatedOrder := newUpdatedOrder()

		useCaseMock := mocks.NewUpdateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, mock.AnythingOfType("*models.Order")).
			Return(updatedOrder, nil)

		handler := NewOrderHandler(nil, nil, nil, nil, useCaseMock, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response responses.UpdateOrderResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		if assert.NotNil(t, response.Order) {
			assert.Equal(t, updatedOrder.ID, response.Order.ID)
			assert.Equal(t, updatedOrder.Total, response.Order.Total)
		}
	})
}

// =============================================================================
// Update Tests - Invalid Request
// =============================================================================

func TestOrderHandler_Update_InvalidRequest(t *testing.T) {
	t.Run("when body is invalid JSON then returns 400", func(t *testing.T) {
		// Arrange
		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when customer name is missing then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()
		request.Order.Customer.Name = ""

		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when items is empty then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()
		request.Order.Items = []requests.OrderItemRequest{}

		handler := NewOrderHandler(nil, nil, nil, nil, nil, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// =============================================================================
// Update Tests - Domain Errors
// =============================================================================

func TestOrderHandler_Update_DomainErrors(t *testing.T) {
	t.Run("when order not found then returns 404", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()

		useCaseMock := mocks.NewUpdateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 999, mock.AnythingOfType("*models.Order")).
			Return(nil, &domainErrors.RecordNotFoundError{Message: domainErrors.OrderNotFound})

		handler := NewOrderHandler(nil, nil, nil, nil, useCaseMock, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "999"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("when order is not editable then returns 422", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()

		useCaseMock := mocks.NewUpdateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, mock.AnythingOfType("*models.Order")).
			Return(nil, &domainErrors.BusinessRuleError{Message: domainErrors.OrderCannotBeEdited})

		handler := NewOrderHandler(nil, nil, nil, nil, useCaseMock, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("when validation error then returns 400", func(t *testing.T) {
		// Arrange
		request := newValidUpdateOrderRequest()

		useCaseMock := mocks.NewUpdateOrderUseCase(t)
		useCaseMock.EXPECT().
			Execute(mock.Anything, 1, 1, mock.AnythingOfType("*models.Order")).
			Return(nil, &domainErrors.ValidationError{Message: domainErrors.SubtotalMismatch})

		handler := NewOrderHandler(nil, nil, nil, nil, useCaseMock, nil)

		body, _ := json.Marshal(request)
		req := httptest.NewRequest(http.MethodPut, "/shops/1/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1", "order_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.Update(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
