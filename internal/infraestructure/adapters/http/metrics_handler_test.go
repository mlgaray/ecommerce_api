package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
	"github.com/mlgaray/ecommerce_api/mocks"
)

// =============================================================================
// Test Helpers - Metrics
// =============================================================================

func newTestDashboard() *models.DashboardMetrics {
	avg := 24.0
	return &models.DashboardMetrics{
		Revenue: models.RevenueMetrics{
			Today: models.RevenuePeriodComparison{
				Current:  models.RevenueSummary{TotalRevenue: 1000, OrderCount: 10, AOV: 100},
				Previous: models.RevenueSummary{TotalRevenue: 800, OrderCount: 8, AOV: 100},
			},
		},
		Orders: models.OrdersMetrics{
			TotalOrders:       10,
			AvgCompletionTime: &avg,
			StatusDistribution: []models.OrderStatusDistribution{
				{Status: "completed", Count: 10},
			},
		},
		TopProducts:          []models.TopProduct{{ProductID: 1, ProductName: "Widget"}},
		PaymentDistribution:  []models.MethodDistribution{{Name: "Cash", OrderCount: 10}},
		DeliveryDistribution: []models.MethodDistribution{{Name: "Delivery", OrderCount: 5}},
		Customers:            models.CustomerOverview{TotalUnique: 50},
	}
}

func newTestTrendPoints() []models.RevenueTrendPoint {
	return []models.RevenueTrendPoint{
		{Date: "2024-06-01", Revenue: 100, OrderCount: 5},
		{Date: "2024-06-02", Revenue: 200, OrderCount: 10},
	}
}

func newTestTopProducts() []models.TopProduct {
	return []models.TopProduct{
		{ProductID: 1, ProductName: "Widget", QuantitySold: 100, Revenue: 999.99},
	}
}

func newTestTopCustomers() []models.TopCustomer {
	return []models.TopCustomer{
		{CustomerName: "John", CustomerPhone: "123", OrderCount: 5, TotalSpent: 500},
	}
}

func newMetricsHandler(
	dashboardMock *mocks.GetDashboardMetricsUseCase,
	trendMock *mocks.GetRevenueTrendUseCase,
	productsMock *mocks.GetTopProductsUseCase,
	customersMock *mocks.GetTopCustomersUseCase,
) *MetricsHandler {
	return &MetricsHandler{
		getDashboardUseCase:    dashboardMock,
		getRevenueTrendUseCase: trendMock,
		getTopProductsUseCase:  productsMock,
		getTopCustomersUseCase: customersMock,
	}
}

// =============================================================================
// GetDashboard Tests
// =============================================================================

func TestMetricsHandler_GetDashboard(t *testing.T) {
	t.Run("when valid request then returns 200 with dashboard", func(t *testing.T) {
		// Arrange
		dashboard := newTestDashboard()

		dashboardMock := mocks.NewGetDashboardMetricsUseCase(t)
		dashboardMock.EXPECT().
			Execute(mock.Anything, 1, (*string)(nil)).
			Return(dashboard, nil)

		handler := newMetricsHandler(dashboardMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response responses.DashboardResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1000.0, response.Revenue.Today.Current.TotalRevenue)
	})

	t.Run("when shop_id is not numeric then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assertErrorMessage(t, rr, "invalid_shop_id_format")
	})

	t.Run("when shop_id is zero then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/0/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "0"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when shop_id is negative then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/-1/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "-1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when tz query param provided then passes it to use case", func(t *testing.T) {
		// Arrange
		dashboard := newTestDashboard()
		tz := "America/New_York"

		dashboardMock := mocks.NewGetDashboardMetricsUseCase(t)
		dashboardMock.EXPECT().
			Execute(mock.Anything, 1, &tz).
			Return(dashboard, nil)

		handler := newMetricsHandler(dashboardMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/dashboard?tz=America/New_York", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when no tz param then passes nil", func(t *testing.T) {
		// Arrange
		dashboard := newTestDashboard()

		dashboardMock := mocks.NewGetDashboardMetricsUseCase(t)
		dashboardMock.EXPECT().
			Execute(mock.Anything, 1, (*string)(nil)).
			Return(dashboard, nil)

		handler := newMetricsHandler(dashboardMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("when use case errors then returns 500", func(t *testing.T) {
		// Arrange
		dashboardMock := mocks.NewGetDashboardMetricsUseCase(t)
		dashboardMock.EXPECT().
			Execute(mock.Anything, 1, (*string)(nil)).
			Return(nil, errors.New("internal error"))

		handler := newMetricsHandler(dashboardMock, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/dashboard", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetDashboard(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// =============================================================================
// GetRevenueTrend Tests
// =============================================================================

func TestMetricsHandler_GetRevenueTrend(t *testing.T) {
	t.Run("when valid request then returns 200", func(t *testing.T) {
		// Arrange
		points := newTestTrendPoints()

		trendMock := mocks.NewGetRevenueTrendUseCase(t)
		trendMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(points, nil)

		handler := newMetricsHandler(nil, trendMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response responses.RevenueTrendResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Trend, 2)
	})

	t.Run("when shop_id is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/metrics/revenue/trend", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when date_from is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend?date_from=invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assertErrorMessage(t, rr, "invalid_date_from_format")
	})

	t.Run("when date_to is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend?date_to=not-a-date", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assertErrorMessage(t, rr, "invalid_date_to_format")
	})

	t.Run("when limit is not numeric then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend?limit=abc", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assertErrorMessage(t, rr, "invalid_limit_format")
	})

	t.Run("when limit is negative then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend?limit=-5", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assertErrorMessage(t, rr, "limit_cannot_be_negative")
	})

	t.Run("when use case errors then returns 500", func(t *testing.T) {
		// Arrange
		trendMock := mocks.NewGetRevenueTrendUseCase(t)
		trendMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(nil, errors.New("internal error"))

		handler := newMetricsHandler(nil, trendMock, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/revenue/trend", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetRevenueTrend(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// =============================================================================
// GetTopProducts Tests
// =============================================================================

func TestMetricsHandler_GetTopProducts(t *testing.T) {
	t.Run("when valid request then returns 200", func(t *testing.T) {
		// Arrange
		products := newTestTopProducts()

		productsMock := mocks.NewGetTopProductsUseCase(t)
		productsMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(products, nil)

		handler := newMetricsHandler(nil, nil, productsMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/products/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response responses.TopProductsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Products, 1)
	})

	t.Run("when shop_id is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/metrics/products/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when invalid filters then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/products/top?date_from=bad", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when use case errors then returns 500", func(t *testing.T) {
		// Arrange
		productsMock := mocks.NewGetTopProductsUseCase(t)
		productsMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(nil, errors.New("internal error"))

		handler := newMetricsHandler(nil, nil, productsMock, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/products/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopProducts(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// =============================================================================
// GetTopCustomers Tests
// =============================================================================

func TestMetricsHandler_GetTopCustomers(t *testing.T) {
	t.Run("when valid request then returns 200", func(t *testing.T) {
		// Arrange
		customers := newTestTopCustomers()

		customersMock := mocks.NewGetTopCustomersUseCase(t)
		customersMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(customers, nil)

		handler := newMetricsHandler(nil, nil, nil, customersMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/customers/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopCustomers(rr, req)

		// Assert
		assert.Equal(t, http.StatusOK, rr.Code)

		var response responses.TopCustomersResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Customers, 1)
		assert.Equal(t, "John", response.Customers[0].CustomerName)
	})

	t.Run("when shop_id is invalid then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/abc/metrics/customers/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "abc"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopCustomers(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when invalid filters then returns 400", func(t *testing.T) {
		// Arrange
		handler := newMetricsHandler(nil, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/customers/top?limit=abc", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopCustomers(rr, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("when use case errors then returns 500", func(t *testing.T) {
		// Arrange
		customersMock := mocks.NewGetTopCustomersUseCase(t)
		customersMock.EXPECT().
			Execute(mock.Anything, 1, mock.AnythingOfType("models.MetricsFilters")).
			Return(nil, errors.New("internal error"))

		handler := newMetricsHandler(nil, nil, nil, customersMock)

		req := httptest.NewRequest(http.MethodGet, "/shops/1/metrics/customers/top", nil)
		req = mux.SetURLVars(req, map[string]string{"shop_id": "1"})
		rr := httptest.NewRecorder()

		// Act
		handler.GetTopCustomers(rr, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

// =============================================================================
// Assertion helper
// =============================================================================

func assertErrorMessage(t *testing.T, rr *httptest.ResponseRecorder, expectedMessage string) {
	t.Helper()
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, errResp["error"])
}
