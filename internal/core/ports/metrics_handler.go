package ports

import "net/http"

// MetricsHandler handles HTTP requests for shop metrics.
type MetricsHandler interface {
	GetDashboard(http.ResponseWriter, *http.Request)
	GetRevenueTrend(http.ResponseWriter, *http.Request)
	GetTopProducts(http.ResponseWriter, *http.Request)
	GetTopCustomers(http.ResponseWriter, *http.Request)
	GetShippingSummary(http.ResponseWriter, *http.Request)
	GetVisitsTrend(http.ResponseWriter, *http.Request)
}
