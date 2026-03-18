package models

import "time"

// RevenuePeriod defines a date range for revenue aggregation.
type RevenuePeriod struct {
	From time.Time
	To   time.Time
}

// RevenueSummary contains revenue aggregation for a single period.
type RevenueSummary struct {
	TotalRevenue float64 `json:"total_revenue"`
	OrderCount   int     `json:"order_count"`
	AOV          float64 `json:"aov"` // Average Order Value: TotalRevenue / OrderCount
}

// RevenuePeriodComparison compares current period vs previous period.
type RevenuePeriodComparison struct {
	Current       RevenueSummary `json:"current"`
	Previous      RevenueSummary `json:"previous"`
	RevenueChange float64        `json:"revenue_change"` // % change
	OrderChange   float64        `json:"order_change"`
	AOVChange     float64        `json:"aov_change"`
}

// RevenueMetrics contains revenue comparisons for today, this week, and this month.
type RevenueMetrics struct {
	Today     RevenuePeriodComparison `json:"today"`
	ThisWeek  RevenuePeriodComparison `json:"this_week"`
	ThisMonth RevenuePeriodComparison `json:"this_month"`
}

// OrderStatusDistribution represents the count of orders per status.
type OrderStatusDistribution struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// OrdersMetrics contains order statistics for the current month.
type OrdersMetrics struct {
	TotalOrders        int                       `json:"total_orders"`
	StatusDistribution []OrderStatusDistribution `json:"status_distribution"`
	CancellationRate   float64                   `json:"cancellation_rate"`
	AvgCompletionTime  *float64                  `json:"avg_completion_time_hours"`
}

// TopProduct represents a product ranked by quantity sold or revenue.
type TopProduct struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ImageURL     string  `json:"image_url,omitempty"`
	QuantitySold int     `json:"quantity_sold"`
	Revenue      float64 `json:"revenue"`
}

// CustomerOverview contains customer aggregation stats.
type CustomerOverview struct {
	TotalUnique int `json:"total_unique"`
	NewCount    int `json:"new_count"`
	Returning   int `json:"returning"`
}

// TopCustomer represents a customer ranked by total spend.
type TopCustomer struct {
	CustomerName  string  `json:"customer_name"`
	CustomerPhone string  `json:"customer_phone"`
	OrderCount    int     `json:"order_count"`
	TotalSpent    float64 `json:"total_spent"`
}

// MethodDistribution represents the distribution of a payment or delivery method.
type MethodDistribution struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	OrderCount int     `json:"order_count"`
	Percentage float64 `json:"percentage"`
}

// RevenueTrendPoint represents a single data point in a revenue time series.
type RevenueTrendPoint struct {
	Date       string  `json:"date"` // "2006-01-02" format
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

// VisitsPeriodComparison compares current visit count vs previous period.
type VisitsPeriodComparison struct {
	Current  int     `json:"current"`
	Previous int     `json:"previous"`
	Change   float64 `json:"change"`
}

// VisitsSummary contains visit comparisons for today, this week, and this month.
type VisitsSummary struct {
	Today     VisitsPeriodComparison `json:"today"`
	ThisWeek  VisitsPeriodComparison `json:"this_week"`
	ThisMonth VisitsPeriodComparison `json:"this_month"`
}

// VisitsTrendPoint represents a single data point in a visits time series.
type VisitsTrendPoint struct {
	Date       string `json:"date"`
	VisitCount int    `json:"visit_count"`
}

// DashboardMetrics is the complete dashboard response.
type DashboardMetrics struct {
	Revenue              RevenueMetrics       `json:"revenue"`
	Orders               OrdersMetrics        `json:"orders"`
	TopProducts          []TopProduct         `json:"top_products"`
	PaymentDistribution  []MethodDistribution `json:"payment_distribution"`
	DeliveryDistribution []MethodDistribution `json:"delivery_distribution"`
	Customers            CustomerOverview     `json:"customers"`
	Visits               VisitsSummary        `json:"visits"`
}

// ShippingSummary contains aggregated shipping metrics for a date range.
type ShippingSummary struct {
	TotalShippingRevenue float64 `json:"total_shipping_revenue"`
	OrderCount           int     `json:"order_count"`
	AverageShippingCost  float64 `json:"average_shipping_cost"`
}

// MetricsFilters contains filter parameters for metrics endpoints.
type MetricsFilters struct {
	DateFrom *time.Time
	DateTo   *time.Time
	Timezone *string
	Limit    int    // default 10, max 50
	SortBy   string // "quantity" | "revenue"
}

// Validated returns a copy of filters with defaults applied and constraints enforced.
func (f MetricsFilters) Validated() MetricsFilters {
	validated := f

	if validated.Limit <= 0 {
		validated.Limit = 10
	}
	if validated.Limit > 50 {
		validated.Limit = 50
	}

	if validated.SortBy != "revenue" {
		validated.SortBy = "quantity"
	}

	return validated
}
