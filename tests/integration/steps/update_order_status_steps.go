package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	scenarioStatusTransition = "status-transition"
	scenarioStatusNotFound   = "status-not-found"
	scenarioConcurrentModify = "concurrent-modification"

	statusStepsItemsJSON = `[{"id":1,"product_id":10,"product_name":"Test Product","product_image_url":"","base_price":100,"is_promotional":false,"promotional_price":0,"quantity":2,"unit_price":100,"total_price":200,"notes":"","selected_options":[]}]`
)

type UpdateOrderStatusSteps struct {
	currentStatus string
}

func NewUpdateOrderStatusSteps() *UpdateOrderStatusSteps {
	return &UpdateOrderStatusSteps{}
}

// ===== Given Steps =====

func (g *UpdateOrderStatusSteps) anOrderInShopHasStatus(orderID int, shopID int, status string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStatusTransition
	g.currentStatus = status
	return nil
}

func (g *UpdateOrderStatusSteps) anOrderDoesNotExistInShopForStatusUpdate(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStatusNotFound
	return nil
}

func (g *UpdateOrderStatusSteps) anOrderInShopHasStatusButWillBeConcurrentlyModified(orderID int, shopID int, status string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioConcurrentModify
	g.currentStatus = status
	return nil
}

// ===== When Steps =====

func (g *UpdateOrderStatusSteps) iSendAnUpdateOrderStatusRequest(shopID int, orderID int, newStatus string) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupUpdateStatusSQLExpectations(ctx, shopID, orderID, newStatus)

	body := map[string]string{"status": newStatus}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/shops/%d/orders/%d/status", ctx.server.URL, shopID, orderID)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseStatusResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *UpdateOrderStatusSteps) setupUpdateStatusSQLExpectations(ctx *TestContext, shopID int, orderID int, newStatus string) {
	getByIDColumns := []string{
		"id", "order_number", "status",
		"store_id", "store_name", "store_slug",
		"customer_name", "customer_phone", "customer_email",
		"customer_address_name", "customer_address_place_id",
		"customer_address_lat", "customer_address_lng",
		"payment_method_id", "payment_method_code", "payment_method_name",
		"delivery_method_id", "delivery_method_code", "delivery_method_name",
		"delivery_zone_id", "delivery_zone_name", "delivery_zone_price",
		"subtotal", "shipping_cost", "total",
		"created_at", "updated_at",
		"items",
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioStatusTransition:
		// Step 1: GetByID - fetch current order
		rows1 := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", g.currentStatus,
				shopID, "Test Store", "test-store",
				"John Doe", "1234567890", "john@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				nil, nil, nil,
				200.0, 50.0, 250.0,
				now, now,
				statusStepsItemsJSON,
			)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows1)

		// Step 2: UPDATE status (optimistic lock)
		ctx.mockSQLMock.ExpectExec("UPDATE orders SET status").
			WithArgs(newStatus, orderID, shopID, g.currentStatus).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Step 3: GetByID - re-fetch with new status
		rows2 := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", newStatus,
				shopID, "Test Store", "test-store",
				"John Doe", "1234567890", "john@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				nil, nil, nil,
				200.0, 50.0, 250.0,
				now, now,
				statusStepsItemsJSON,
			)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows2)

	case scenarioStatusNotFound:
		emptyRows := sqlmock.NewRows(getByIDColumns)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(emptyRows)

	case scenarioConcurrentModify:
		// Step 1: GetByID - fetch current order
		rows1 := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", g.currentStatus,
				shopID, "Test Store", "test-store",
				"John Doe", "1234567890", "john@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				nil, nil, nil,
				200.0, 50.0, 250.0,
				now, now,
				statusStepsItemsJSON,
			)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows1)

		// Step 2: UPDATE returns 0 rows affected (concurrent modification)
		ctx.mockSQLMock.ExpectExec("UPDATE orders SET status").
			WithArgs(newStatus, orderID, shopID, g.currentStatus).
			WillReturnResult(sqlmock.NewResult(0, 0))
	}
}

func (g *UpdateOrderStatusSteps) parseStatusResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResponse map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			ctx.errorMessage = errorResponse["error"]
		}
	} else {
		var response responses.UpdateOrderStatusResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response.Order
		}
	}
}

// ===== Then Steps =====

func (g *UpdateOrderStatusSteps) theResponseShouldContainTheOrderWithStatus(expectedStatus string) error {
	ctx := GetTestContext()
	order, ok := ctx.responseBody.(*responses.OrderResponse)
	if !ok {
		return fmt.Errorf("expected OrderResponse, got: %T", ctx.responseBody)
	}
	if order == nil {
		return fmt.Errorf("expected order in response, got nil")
	}
	if order.Status != expectedStatus {
		return fmt.Errorf("expected order status '%s', got '%s'", expectedStatus, order.Status)
	}
	return nil
}

// ===== Register Steps =====

func (g *UpdateOrderStatusSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^an order with ID (\d+) in shop (\d+) has status "([^"]*)"$`, g.anOrderInShopHasStatus)
	sc.Step(`^an order with ID (\d+) does not exist in shop (\d+) for status update$`, g.anOrderDoesNotExistInShopForStatusUpdate)
	sc.Step(`^an order with ID (\d+) in shop (\d+) has status "([^"]*)" but will be concurrently modified$`, g.anOrderInShopHasStatusButWillBeConcurrentlyModified)

	// When steps
	sc.Step(`^I send an update order status request for shop (\d+) and order (\d+) with status "([^"]*)"$`, g.iSendAnUpdateOrderStatusRequest)

	// Then steps
	sc.Step(`^the response should contain the order with status "([^"]*)"$`, g.theResponseShouldContainTheOrderWithStatus)
}
