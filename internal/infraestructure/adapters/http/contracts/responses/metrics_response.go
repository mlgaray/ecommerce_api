package responses

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// DashboardResponse represents the HTTP response for the dashboard endpoint.
type DashboardResponse struct {
	Revenue              RevenueMetricsResponse   `json:"revenue"`
	Orders               OrdersMetricsResponse    `json:"orders"`
	TopProducts          []TopProductResponse     `json:"top_products"`
	PaymentDistribution  []MethodDistResponse     `json:"payment_distribution"`
	DeliveryDistribution []MethodDistResponse     `json:"delivery_distribution"`
	Customers            CustomerOverviewResponse `json:"customers"`
	Visits               VisitsSummaryResponse    `json:"visits"`
}

// VisitsSummaryResponse represents visit comparisons for the dashboard.
type VisitsSummaryResponse struct {
	Today     VisitsPeriodResponse `json:"today"`
	ThisWeek  VisitsPeriodResponse `json:"this_week"`
	ThisMonth VisitsPeriodResponse `json:"this_month"`
}

// VisitsPeriodResponse represents a visit period comparison.
type VisitsPeriodResponse struct {
	Current  int     `json:"current"`
	Previous int     `json:"previous"`
	Change   float64 `json:"change"`
}

// VisitsTrendPointResponse represents a single data point in the visits trend.
type VisitsTrendPointResponse struct {
	Date       string `json:"date"`
	VisitCount int    `json:"visit_count"`
}

// VisitsTrendResponse is the HTTP response for the visits trend endpoint.
type VisitsTrendResponse struct {
	Trend []VisitsTrendPointResponse `json:"trend"`
}

// RevenueMetricsResponse represents revenue comparisons.
type RevenueMetricsResponse struct {
	Today     RevenuePeriodResponse `json:"today"`
	ThisWeek  RevenuePeriodResponse `json:"this_week"`
	ThisMonth RevenuePeriodResponse `json:"this_month"`
}

// RevenuePeriodResponse represents a period comparison.
type RevenuePeriodResponse struct {
	Current       RevenueSummaryResponse `json:"current"`
	Previous      RevenueSummaryResponse `json:"previous"`
	RevenueChange float64                `json:"revenue_change"`
	OrderChange   float64                `json:"order_change"`
	AOVChange     float64                `json:"aov_change"`
}

// RevenueSummaryResponse represents revenue aggregation for a period.
type RevenueSummaryResponse struct {
	TotalRevenue float64 `json:"total_revenue"`
	OrderCount   int     `json:"order_count"`
	AOV          float64 `json:"aov"`
}

// OrdersMetricsResponse represents order statistics.
type OrdersMetricsResponse struct {
	TotalOrders        int                       `json:"total_orders"`
	StatusDistribution []OrderStatusDistResponse `json:"status_distribution"`
	CancellationRate   float64                   `json:"cancellation_rate"`
	AvgCompletionTime  *float64                  `json:"avg_completion_time_hours"`
}

// OrderStatusDistResponse represents the count per status.
type OrderStatusDistResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// TopProductResponse represents a top product.
type TopProductResponse struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ImageURL     string  `json:"image_url,omitempty"`
	QuantitySold int     `json:"quantity_sold"`
	Revenue      float64 `json:"revenue"`
}

// CustomerOverviewResponse represents customer stats.
type CustomerOverviewResponse struct {
	TotalUnique int `json:"total_unique"`
	NewCount    int `json:"new_count"`
	Returning   int `json:"returning"`
}

// TopCustomerResponse represents a top customer.
type TopCustomerResponse struct {
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	OrderCount    int     `json:"order_count"`
	TotalSpent    float64 `json:"total_spent"`
}

// MethodDistResponse represents a payment/delivery method distribution entry.
type MethodDistResponse struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	OrderCount int     `json:"order_count"`
	Percentage float64 `json:"percentage"`
}

// RevenueTrendPointResponse represents a single data point in the trend.
type RevenueTrendPointResponse struct {
	Date       string  `json:"date"`
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

// RevenueTrendResponse is the HTTP response for the revenue trend endpoint.
type RevenueTrendResponse struct {
	Trend []RevenueTrendPointResponse `json:"trend"`
}

// TopProductsResponse is the HTTP response for the top products endpoint.
type TopProductsResponse struct {
	Products []TopProductResponse `json:"products"`
}

// TopCustomersResponse is the HTTP response for the top customers endpoint.
type TopCustomersResponse struct {
	Customers []TopCustomerResponse `json:"customers"`
}

// ShippingSummaryResponse represents the HTTP response for the shipping summary endpoint.
type ShippingSummaryResponse struct {
	TotalShippingRevenue float64 `json:"total_shipping_revenue"`
	OrderCount           int     `json:"order_count"`
	AverageShippingCost  float64 `json:"average_shipping_cost"`
}

// NewShippingSummaryResponse converts a domain ShippingSummary to an HTTP response.
func NewShippingSummaryResponse(s models.ShippingSummary) *ShippingSummaryResponse {
	return &ShippingSummaryResponse{
		TotalShippingRevenue: s.TotalShippingRevenue,
		OrderCount:           s.OrderCount,
		AverageShippingCost:  s.AverageShippingCost,
	}
}

// NewDashboardResponse converts a domain DashboardMetrics to an HTTP response.
func NewDashboardResponse(d *models.DashboardMetrics) *DashboardResponse {
	response := &DashboardResponse{
		Revenue: RevenueMetricsResponse{
			Today:     newRevenuePeriodResponse(d.Revenue.Today),
			ThisWeek:  newRevenuePeriodResponse(d.Revenue.ThisWeek),
			ThisMonth: newRevenuePeriodResponse(d.Revenue.ThisMonth),
		},
		Orders: OrdersMetricsResponse{
			TotalOrders:       d.Orders.TotalOrders,
			CancellationRate:  d.Orders.CancellationRate,
			AvgCompletionTime: d.Orders.AvgCompletionTime,
		},
		Customers: CustomerOverviewResponse{
			TotalUnique: d.Customers.TotalUnique,
			NewCount:    d.Customers.NewCount,
			Returning:   d.Customers.Returning,
		},
	}

	// Status distribution
	response.Orders.StatusDistribution = make([]OrderStatusDistResponse, len(d.Orders.StatusDistribution))
	for i, s := range d.Orders.StatusDistribution {
		response.Orders.StatusDistribution[i] = OrderStatusDistResponse{
			Status: s.Status,
			Count:  s.Count,
		}
	}

	// Top products
	response.TopProducts = make([]TopProductResponse, len(d.TopProducts))
	for i, p := range d.TopProducts {
		response.TopProducts[i] = TopProductResponse{
			ProductID:    p.ProductID,
			ProductName:  p.ProductName,
			ImageURL:     p.ImageURL,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		}
	}

	// Payment distribution
	response.PaymentDistribution = make([]MethodDistResponse, len(d.PaymentDistribution))
	for i, m := range d.PaymentDistribution {
		response.PaymentDistribution[i] = MethodDistResponse{
			Name:       m.Name,
			Code:       m.Code,
			OrderCount: m.OrderCount,
			Percentage: m.Percentage,
		}
	}

	// Delivery distribution
	response.DeliveryDistribution = make([]MethodDistResponse, len(d.DeliveryDistribution))
	for i, m := range d.DeliveryDistribution {
		response.DeliveryDistribution[i] = MethodDistResponse{
			Name:       m.Name,
			Code:       m.Code,
			OrderCount: m.OrderCount,
			Percentage: m.Percentage,
		}
	}

	// Visits
	response.Visits = VisitsSummaryResponse{
		Today:     newVisitsPeriodResponse(d.Visits.Today),
		ThisWeek:  newVisitsPeriodResponse(d.Visits.ThisWeek),
		ThisMonth: newVisitsPeriodResponse(d.Visits.ThisMonth),
	}

	return response
}

func newVisitsPeriodResponse(p models.VisitsPeriodComparison) VisitsPeriodResponse {
	return VisitsPeriodResponse{
		Current:  p.Current,
		Previous: p.Previous,
		Change:   p.Change,
	}
}

func newRevenuePeriodResponse(p models.RevenuePeriodComparison) RevenuePeriodResponse {
	return RevenuePeriodResponse{
		Current: RevenueSummaryResponse{
			TotalRevenue: p.Current.TotalRevenue,
			OrderCount:   p.Current.OrderCount,
			AOV:          p.Current.AOV,
		},
		Previous: RevenueSummaryResponse{
			TotalRevenue: p.Previous.TotalRevenue,
			OrderCount:   p.Previous.OrderCount,
			AOV:          p.Previous.AOV,
		},
		RevenueChange: p.RevenueChange,
		OrderChange:   p.OrderChange,
		AOVChange:     p.AOVChange,
	}
}

// NewRevenueTrendResponse converts domain trend points to an HTTP response.
func NewRevenueTrendResponse(points []models.RevenueTrendPoint) *RevenueTrendResponse {
	trend := make([]RevenueTrendPointResponse, len(points))
	for i, p := range points {
		trend[i] = RevenueTrendPointResponse{
			Date:       p.Date,
			Revenue:    p.Revenue,
			OrderCount: p.OrderCount,
		}
	}
	return &RevenueTrendResponse{Trend: trend}
}

// NewTopProductsResponse converts domain top products to an HTTP response.
func NewTopProductsResponse(products []models.TopProduct) *TopProductsResponse {
	resp := make([]TopProductResponse, len(products))
	for i, p := range products {
		resp[i] = TopProductResponse{
			ProductID:    p.ProductID,
			ProductName:  p.ProductName,
			ImageURL:     p.ImageURL,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		}
	}
	return &TopProductsResponse{Products: resp}
}

// NewVisitsTrendResponse converts domain visits trend points to an HTTP response.
func NewVisitsTrendResponse(points []models.VisitsTrendPoint) *VisitsTrendResponse {
	trend := make([]VisitsTrendPointResponse, len(points))
	for i, p := range points {
		trend[i] = VisitsTrendPointResponse{
			Date:       p.Date,
			VisitCount: p.VisitCount,
		}
	}
	return &VisitsTrendResponse{Trend: trend}
}

// NewTopCustomersResponse converts domain top customers to an HTTP response.
func NewTopCustomersResponse(customers []models.TopCustomer) *TopCustomersResponse {
	resp := make([]TopCustomerResponse, len(customers))
	for i, c := range customers {
		resp[i] = TopCustomerResponse{
			CustomerName:  c.CustomerName,
			CustomerPhone: c.CustomerPhone,
			OrderCount:    c.OrderCount,
			TotalSpent:    c.TotalSpent,
		}
	}
	return &TopCustomersResponse{Customers: resp}
}
