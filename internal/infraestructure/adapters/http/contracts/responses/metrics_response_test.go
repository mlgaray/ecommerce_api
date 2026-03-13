package responses

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newTestDashboardMetrics() *models.DashboardMetrics {
	avgTime := 2.5
	return &models.DashboardMetrics{
		Revenue: models.RevenueMetrics{
			Today: models.RevenuePeriodComparison{
				Current:       models.RevenueSummary{TotalRevenue: 1000, OrderCount: 10, AOV: 100},
				Previous:      models.RevenueSummary{TotalRevenue: 800, OrderCount: 8, AOV: 100},
				RevenueChange: 25.0,
				OrderChange:   25.0,
				AOVChange:     0.0,
			},
			ThisWeek: models.RevenuePeriodComparison{
				Current:       models.RevenueSummary{TotalRevenue: 5000, OrderCount: 50, AOV: 100},
				Previous:      models.RevenueSummary{TotalRevenue: 4000, OrderCount: 40, AOV: 100},
				RevenueChange: 25.0,
				OrderChange:   25.0,
				AOVChange:     0.0,
			},
			ThisMonth: models.RevenuePeriodComparison{
				Current:       models.RevenueSummary{TotalRevenue: 20000, OrderCount: 200, AOV: 100},
				Previous:      models.RevenueSummary{TotalRevenue: 18000, OrderCount: 180, AOV: 100},
				RevenueChange: 11.11,
				OrderChange:   11.11,
				AOVChange:     0.0,
			},
		},
		Orders: models.OrdersMetrics{
			TotalOrders:       200,
			CancellationRate:  5.0,
			AvgCompletionTime: &avgTime,
			StatusDistribution: []models.OrderStatusDistribution{
				{Status: "pending", Count: 50},
				{Status: "completed", Count: 120},
			},
		},
		TopProducts: []models.TopProduct{
			{ProductID: 1, ProductName: "Laptop", ImageURL: "https://img.com/1.jpg", QuantitySold: 50, Revenue: 49999.50},
		},
		PaymentDistribution: []models.MethodDistribution{
			{Name: "Cash", Code: "cash", OrderCount: 100, Percentage: 50.0},
		},
		DeliveryDistribution: []models.MethodDistribution{
			{Name: "Delivery", Code: "delivery", OrderCount: 150, Percentage: 75.0},
		},
		Customers: models.CustomerOverview{
			TotalUnique: 180,
			NewCount:    30,
			Returning:   150,
		},
	}
}

// =============================================================================
// NewShippingSummaryResponse
// =============================================================================

func TestNewShippingSummaryResponse(t *testing.T) {
	t.Run("when valid model then maps all fields", func(t *testing.T) {
		model := models.ShippingSummary{
			TotalShippingRevenue: 15000.50,
			OrderCount:           300,
			AverageShippingCost:  50.0,
		}

		resp := NewShippingSummaryResponse(model)

		assert.Equal(t, 15000.50, resp.TotalShippingRevenue)
		assert.Equal(t, 300, resp.OrderCount)
		assert.Equal(t, 50.0, resp.AverageShippingCost)
	})

	t.Run("when zero values then maps zeros", func(t *testing.T) {
		resp := NewShippingSummaryResponse(models.ShippingSummary{})

		assert.Equal(t, 0.0, resp.TotalShippingRevenue)
		assert.Equal(t, 0, resp.OrderCount)
		assert.Equal(t, 0.0, resp.AverageShippingCost)
	})
}

// =============================================================================
// NewDashboardResponse
// =============================================================================

func TestNewDashboardResponse(t *testing.T) {
	t.Run("when full model then maps revenue periods", func(t *testing.T) {
		dashboard := newTestDashboardMetrics()

		resp := NewDashboardResponse(dashboard)

		assert.Equal(t, 1000.0, resp.Revenue.Today.Current.TotalRevenue)
		assert.Equal(t, 10, resp.Revenue.Today.Current.OrderCount)
		assert.Equal(t, 800.0, resp.Revenue.Today.Previous.TotalRevenue)
		assert.Equal(t, 25.0, resp.Revenue.Today.RevenueChange)

		assert.Equal(t, 5000.0, resp.Revenue.ThisWeek.Current.TotalRevenue)
		assert.Equal(t, 20000.0, resp.Revenue.ThisMonth.Current.TotalRevenue)
	})

	t.Run("when full model then maps orders metrics", func(t *testing.T) {
		dashboard := newTestDashboardMetrics()

		resp := NewDashboardResponse(dashboard)

		assert.Equal(t, 200, resp.Orders.TotalOrders)
		assert.Equal(t, 5.0, resp.Orders.CancellationRate)
		assert.Equal(t, 2.5, *resp.Orders.AvgCompletionTime)
		assert.Len(t, resp.Orders.StatusDistribution, 2)
		assert.Equal(t, "pending", resp.Orders.StatusDistribution[0].Status)
		assert.Equal(t, 50, resp.Orders.StatusDistribution[0].Count)
	})

	t.Run("when full model then maps top products and distributions", func(t *testing.T) {
		dashboard := newTestDashboardMetrics()

		resp := NewDashboardResponse(dashboard)

		assert.Len(t, resp.TopProducts, 1)
		assert.Equal(t, 1, resp.TopProducts[0].ProductID)
		assert.Equal(t, "Laptop", resp.TopProducts[0].ProductName)
		assert.Equal(t, "https://img.com/1.jpg", resp.TopProducts[0].ImageURL)

		assert.Len(t, resp.PaymentDistribution, 1)
		assert.Equal(t, "Cash", resp.PaymentDistribution[0].Name)
		assert.Equal(t, 50.0, resp.PaymentDistribution[0].Percentage)

		assert.Len(t, resp.DeliveryDistribution, 1)
		assert.Equal(t, "Delivery", resp.DeliveryDistribution[0].Name)
	})

	t.Run("when full model then maps customer overview", func(t *testing.T) {
		dashboard := newTestDashboardMetrics()

		resp := NewDashboardResponse(dashboard)

		assert.Equal(t, 180, resp.Customers.TotalUnique)
		assert.Equal(t, 30, resp.Customers.NewCount)
		assert.Equal(t, 150, resp.Customers.Returning)
	})

	t.Run("when empty slices then response slices are not nil", func(t *testing.T) {
		dashboard := &models.DashboardMetrics{
			TopProducts:          []models.TopProduct{},
			PaymentDistribution:  []models.MethodDistribution{},
			DeliveryDistribution: []models.MethodDistribution{},
			Orders: models.OrdersMetrics{
				StatusDistribution: []models.OrderStatusDistribution{},
			},
		}

		resp := NewDashboardResponse(dashboard)

		assert.NotNil(t, resp.TopProducts)
		assert.Empty(t, resp.TopProducts)
		assert.NotNil(t, resp.PaymentDistribution)
		assert.NotNil(t, resp.DeliveryDistribution)
		assert.NotNil(t, resp.Orders.StatusDistribution)
	})
}

// =============================================================================
// NewRevenueTrendResponse
// =============================================================================

func TestNewRevenueTrendResponse(t *testing.T) {
	t.Run("when points provided then maps all fields", func(t *testing.T) {
		points := []models.RevenueTrendPoint{
			{Date: "2024-01-01", Revenue: 1000, OrderCount: 10},
			{Date: "2024-01-02", Revenue: 1500, OrderCount: 15},
		}

		resp := NewRevenueTrendResponse(points)

		assert.Len(t, resp.Trend, 2)
		assert.Equal(t, "2024-01-01", resp.Trend[0].Date)
		assert.Equal(t, 1000.0, resp.Trend[0].Revenue)
		assert.Equal(t, 10, resp.Trend[0].OrderCount)
	})

	t.Run("when empty points then returns empty trend", func(t *testing.T) {
		resp := NewRevenueTrendResponse([]models.RevenueTrendPoint{})

		assert.NotNil(t, resp.Trend)
		assert.Empty(t, resp.Trend)
	})
}

// =============================================================================
// NewTopProductsResponse
// =============================================================================

func TestNewTopProductsResponse(t *testing.T) {
	t.Run("when products provided then maps all fields", func(t *testing.T) {
		products := []models.TopProduct{
			{ProductID: 1, ProductName: "Laptop", ImageURL: "img.jpg", QuantitySold: 50, Revenue: 49999},
			{ProductID: 2, ProductName: "Mouse", QuantitySold: 200, Revenue: 5000},
		}

		resp := NewTopProductsResponse(products)

		assert.Len(t, resp.Products, 2)
		assert.Equal(t, "Laptop", resp.Products[0].ProductName)
		assert.Equal(t, "img.jpg", resp.Products[0].ImageURL)
		assert.Equal(t, "", resp.Products[1].ImageURL)
	})
}

// =============================================================================
// NewTopCustomersResponse
// =============================================================================

func TestNewTopCustomersResponse(t *testing.T) {
	t.Run("when customers provided then maps all fields", func(t *testing.T) {
		customers := []models.TopCustomer{
			{CustomerName: "John", CustomerPhone: "+54111234", OrderCount: 10, TotalSpent: 5000},
			{CustomerName: "Jane", CustomerPhone: "+54115678", OrderCount: 5, TotalSpent: 3000},
		}

		resp := NewTopCustomersResponse(customers)

		assert.Len(t, resp.Customers, 2)
		assert.Equal(t, "John", resp.Customers[0].CustomerName)
		assert.Equal(t, "+54111234", resp.Customers[0].CustomerPhone)
		assert.Equal(t, 10, resp.Customers[0].OrderCount)
		assert.Equal(t, 5000.0, resp.Customers[0].TotalSpent)
	})
}
