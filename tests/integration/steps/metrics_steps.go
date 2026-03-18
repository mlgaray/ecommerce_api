package steps

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

// =============================================================================
// MetricsSteps
// =============================================================================

type MetricsSteps struct{}

func NewMetricsSteps() *MetricsSteps {
	return &MetricsSteps{}
}

// ===== Given Steps =====

func (m *MetricsSteps) theUserIsAuthenticatedForMetricsOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupMetricsTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetricsSteps) theShopHasMetricsData(shopID int) error {
	ctx := GetTestContext()
	m.setupDashboardSQLExpectations(ctx, shopID)
	return nil
}

func (m *MetricsSteps) theShopHasRevenueTrendData(shopID int) error {
	ctx := GetTestContext()
	m.setupRevenueTrendSQLExpectations(ctx, shopID)
	return nil
}

func (m *MetricsSteps) theShopHasTopProductsData(shopID int) error {
	ctx := GetTestContext()
	m.setupTopProductsSQLExpectations(ctx, shopID)
	return nil
}

func (m *MetricsSteps) theShopHasTopCustomersData(shopID int) error {
	ctx := GetTestContext()
	m.setupTopCustomersSQLExpectations(ctx, shopID)
	return nil
}

// ===== When Steps =====

func (m *MetricsSteps) iSendAGETRequestTo(path string) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupMetricsTestApp(); err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s%s", ctx.server.URL, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	m.parseMetricsResponse(ctx, resp)

	return nil
}

// ===== Then Steps =====

func (m *MetricsSteps) theResponseShouldContainDashboardMetrics() error {
	ctx := GetTestContext()
	dashboard, ok := ctx.responseBody.(*responses.DashboardResponse)
	if !ok {
		return fmt.Errorf("expected DashboardResponse, got: %T", ctx.responseBody)
	}
	if dashboard == nil {
		return fmt.Errorf("expected non-nil dashboard response")
	}
	return nil
}

func (m *MetricsSteps) theResponseShouldContainRevenueTrendData() error {
	ctx := GetTestContext()
	trend, ok := ctx.responseBody.(*responses.RevenueTrendResponse)
	if !ok {
		return fmt.Errorf("expected RevenueTrendResponse, got: %T", ctx.responseBody)
	}
	if trend == nil {
		return fmt.Errorf("expected non-nil revenue trend response")
	}
	return nil
}

func (m *MetricsSteps) theResponseShouldContainTopProductsData() error {
	ctx := GetTestContext()
	products, ok := ctx.responseBody.(*responses.TopProductsResponse)
	if !ok {
		return fmt.Errorf("expected TopProductsResponse, got: %T", ctx.responseBody)
	}
	if products == nil {
		return fmt.Errorf("expected non-nil top products response")
	}
	return nil
}

func (m *MetricsSteps) theResponseShouldContainTopCustomersData() error {
	ctx := GetTestContext()
	customers, ok := ctx.responseBody.(*responses.TopCustomersResponse)
	if !ok {
		return fmt.Errorf("expected TopCustomersResponse, got: %T", ctx.responseBody)
	}
	if customers == nil {
		return fmt.Errorf("expected non-nil top customers response")
	}
	return nil
}

// ===== Response Parser =====

func (m *MetricsSteps) parseMetricsResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		m.parseErrorResponse(ctx, resp)
		return
	}

	bodyBytes := readBody(resp)
	if bodyBytes == nil {
		return
	}

	ctx.responseBody = unmarshalMetricsResponse(bodyBytes)
}

func (m *MetricsSteps) parseErrorResponse(ctx *TestContext, resp *http.Response) {
	var errorResponse map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
		ctx.errorMessage = errorResponse["error"]
	}
}

func readBody(resp *http.Response) []byte {
	bodyBytes := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		bodyBytes = append(bodyBytes, buf[:n]...)
		if err != nil {
			break
		}
	}
	if len(bodyBytes) == 0 {
		return nil
	}
	return bodyBytes
}

func unmarshalMetricsResponse(bodyBytes []byte) interface{} {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil
	}

	if _, ok := raw["revenue"]; ok {
		var dashboard responses.DashboardResponse
		if err := json.Unmarshal(bodyBytes, &dashboard); err == nil {
			return &dashboard
		}
	} else if _, ok := raw["trend"]; ok {
		var trend responses.RevenueTrendResponse
		if err := json.Unmarshal(bodyBytes, &trend); err == nil {
			return &trend
		}
	} else if _, ok := raw["products"]; ok {
		var products responses.TopProductsResponse
		if err := json.Unmarshal(bodyBytes, &products); err == nil {
			return &products
		}
	} else if _, ok := raw["customers"]; ok {
		var customers responses.TopCustomersResponse
		if err := json.Unmarshal(bodyBytes, &customers); err == nil {
			return &customers
		}
	}
	return nil
}

// ===== SQL Mock Setup =====

func (m *MetricsSteps) setupDashboardSQLExpectations(ctx *TestContext, shopID int) {
	// Revenue batch query (returns 12 columns: 2 per period x 6 periods)
	revenueCols := make([]string, 12)
	for i := 0; i < 12; i++ {
		revenueCols[i] = fmt.Sprintf("col%d", i)
	}
	revenueRows := sqlmock.NewRows(revenueCols).AddRow(
		1000.0, 10, // today
		800.0, 8, // yesterday
		5000.0, 50, // this week
		4000.0, 40, // prev week
		20000.0, 200, // this month
		18000.0, 180, // prev month
	)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM orders WHERE store_id").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(revenueRows)

	// Order status distribution
	statusCols := []string{"status", "count"}
	statusRows := sqlmock.NewRows(statusCols).
		AddRow("completed", 150).
		AddRow("pending", 30).
		AddRow("cancelled", 20)
	ctx.mockSQLMock.ExpectQuery("SELECT status, COUNT(.+) FROM orders").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(statusRows)

	// Avg completion time
	avgComp := 24.5
	avgCompRows := sqlmock.NewRows([]string{"avg"}).AddRow(avgComp)
	ctx.mockSQLMock.ExpectQuery("SELECT AVG(.+) FROM orders").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(avgCompRows)

	// Top products (for dashboard - top 5 by quantity)
	topProductCols := []string{"product_id", "product_name", "product_image_url", "quantity_sold", "revenue"}
	topProductRows := sqlmock.NewRows(topProductCols).
		AddRow(1, "Widget", sql.NullString{String: "https://example.com/widget.jpg", Valid: true}, 100, 999.99).
		AddRow(2, "Gadget", sql.NullString{String: "", Valid: false}, 80, 799.99)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM order_items").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), 5).
		WillReturnRows(topProductRows)

	// Payment method distribution
	paymentCols := []string{"payment_method_name", "payment_method_code", "count"}
	paymentRows := sqlmock.NewRows(paymentCols).
		AddRow("Cash", "cash", 120).
		AddRow("Credit Card", "credit_card", 80)
	ctx.mockSQLMock.ExpectQuery("SELECT payment_method_name").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(paymentRows)

	// Delivery method distribution
	deliveryCols := []string{"delivery_method_name", "delivery_method_code", "count"}
	deliveryRows := sqlmock.NewRows(deliveryCols).
		AddRow("Home Delivery", "home_delivery", 150).
		AddRow("Pickup", "pickup", 50)
	ctx.mockSQLMock.ExpectQuery("SELECT delivery_method_name").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(deliveryRows)

	// Customer overview
	customerCols := []string{"total_unique", "new_count", "returning"}
	customerRows := sqlmock.NewRows(customerCols).AddRow(50, 30, 20)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM (.+)customer_orders").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(customerRows)

	// Visits batch query (returns 6 columns: 1 count per period x 6 periods)
	visitsCols := make([]string, 6)
	for i := 0; i < 6; i++ {
		visitsCols[i] = fmt.Sprintf("col%d", i)
	}
	visitsRows := sqlmock.NewRows(visitsCols).AddRow(
		15, 12, // today, yesterday
		85, 70, // this week, prev week
		340, 280, // this month, prev month
	)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM store_visits WHERE shop_id").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(visitsRows)
}

func (m *MetricsSteps) setupRevenueTrendSQLExpectations(ctx *TestContext, shopID int) {
	trendCols := []string{"day", "revenue", "order_count"}
	now := time.Now()
	trendRows := sqlmock.NewRows(trendCols).
		AddRow(now.AddDate(0, 0, -2), 500.0, 5).
		AddRow(now.AddDate(0, 0, -1), 750.0, 8).
		AddRow(now, 1000.0, 10)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM orders WHERE store_id").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(trendRows)
}

func (m *MetricsSteps) setupTopProductsSQLExpectations(ctx *TestContext, shopID int) {
	topProductCols := []string{"product_id", "product_name", "product_image_url", "quantity_sold", "revenue"}
	topProductRows := sqlmock.NewRows(topProductCols).
		AddRow(1, "Widget", sql.NullString{String: "https://example.com/widget.jpg", Valid: true}, 100, 999.99).
		AddRow(2, "Gadget", sql.NullString{String: "", Valid: false}, 80, 799.99)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM order_items").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(topProductRows)
}

func (m *MetricsSteps) setupTopCustomersSQLExpectations(ctx *TestContext, shopID int) {
	customerCols := []string{"customer_name", "customer_phone", "order_count", "total_spent"}
	customerRows := sqlmock.NewRows(customerCols).
		AddRow("John Doe", "123456789", 5, 500.0).
		AddRow("Jane Smith", "987654321", 3, 300.0)
	ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM orders WHERE store_id").
		WithArgs(shopID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(customerRows)
}

// ===== RegisterSteps =====

func (m *MetricsSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given
	sc.Step(`^the user is authenticated for metrics operations$`, m.theUserIsAuthenticatedForMetricsOperations)
	sc.Step(`^the shop (\d+) has metrics data$`, m.theShopHasMetricsData)
	sc.Step(`^the shop (\d+) has revenue trend data$`, m.theShopHasRevenueTrendData)
	sc.Step(`^the shop (\d+) has top products data$`, m.theShopHasTopProductsData)
	sc.Step(`^the shop (\d+) has top customers data$`, m.theShopHasTopCustomersData)

	// When
	sc.Step(`^I send a GET request to "([^"]*)"$`, m.iSendAGETRequestTo)

	// Then
	sc.Step(`^the response should contain dashboard metrics$`, m.theResponseShouldContainDashboardMetrics)
	sc.Step(`^the response should contain revenue trend data$`, m.theResponseShouldContainRevenueTrendData)
	sc.Step(`^the response should contain top products data$`, m.theResponseShouldContainTopProductsData)
	sc.Step(`^the response should contain top customers data$`, m.theResponseShouldContainTopCustomersData)
}
