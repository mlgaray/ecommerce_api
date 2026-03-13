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

const (
	scenarioOrderExists   = "order-exists"
	scenarioOrderNotFound = "order-not-found"
)

type GetOrderByIDSteps struct{}

func NewGetOrderByIDSteps() *GetOrderByIDSteps {
	return &GetOrderByIDSteps{}
}

// ===== Given Steps =====

func (g *GetOrderByIDSteps) theUserIsAuthenticatedForOrderOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GetOrderByIDSteps) anOrderWithIDExistsInShop(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioOrderExists
	return nil
}

func (g *GetOrderByIDSteps) anOrderWithIDDoesNotExistInShop(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioOrderNotFound
	return nil
}

// ===== When Steps =====

func (g *GetOrderByIDSteps) iSendAGetOrderByIDRequestForShopAndOrder(shopID int, orderID int) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetByIDSQLExpectations(ctx, shopID, orderID)

	url := fmt.Sprintf("%s/shops/%d/orders/%d", ctx.server.URL, shopID, orderID)
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
	g.parseOrderResponse(ctx, resp)

	return nil
}

func (g *GetOrderByIDSteps) iSendAGetOrderByIDRequestWithInvalidShopIDAndOrder(invalidShopID string, orderID int) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/shops/%s/orders/%d", ctx.server.URL, invalidShopID, orderID)
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
	g.parseOrderResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *GetOrderByIDSteps) setupGetByIDSQLExpectations(ctx *TestContext, shopID int, orderID int) {
	// 34 columns returned by GetByID CTE query
	columns := []string{
		"id", "order_number", "status",
		"store_id", "store_name", "store_slug",
		"customer_name", "customer_phone", "customer_email",
		"customer_address_name", "customer_address_place_id",
		"customer_address_lat", "customer_address_lng",
		"payment_method_id", "payment_method_code", "payment_method_name",
		"delivery_method_id", "delivery_method_code", "delivery_method_name",
		"delivery_zone_id", "delivery_zone_name", "delivery_zone_price",
		"subtotal", "shipping_cost", "discount", "total",
		"coupon_code", "coupon_type", "coupon_value",
		"coupon_discount_amount", "coupon_min_order_amount",
		"created_at", "updated_at",
		"items",
	}

	switch ctx.scenario {
	case scenarioOrderExists:
		itemsJSON := `[{"id":1,"product_id":10,"product_name":"Test Product","product_image_url":"https://img.com/p.jpg","base_price":100,"is_promotional":false,"promotional_price":0,"quantity":2,"unit_price":100,"total_price":200,"notes":"","selected_options":[]}]`

		now := time.Now()
		rows := sqlmock.NewRows(columns).
			AddRow(
				orderID, "ORD-001", "pending",
				shopID, "Test Store", "test-store",
				"John Doe", "1234567890", "john@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				nil, nil, nil,
				200.0, 50.0, 0.0, 250.0,
				nil, nil, nil, nil, nil,
				now, now,
				itemsJSON,
			)

		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows)

	case scenarioOrderNotFound:
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(emptyRows)
	}
}

func (g *GetOrderByIDSteps) parseOrderResponse(ctx *TestContext, resp *http.Response) {
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
		var response responses.GetOrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response.Order
		}
	}
}

// ===== Then Steps =====

func (g *GetOrderByIDSteps) theResponseShouldContainTheOrderDetails() error {
	ctx := GetTestContext()
	order, ok := ctx.responseBody.(*responses.OrderResponse)
	if !ok {
		return fmt.Errorf("expected OrderResponse, got: %T", ctx.responseBody)
	}
	if order == nil || order.ID == 0 {
		return fmt.Errorf("expected order with ID, got ID 0")
	}
	if order.OrderNumber == "" {
		return fmt.Errorf("expected order with order_number, got empty")
	}
	if order.Status == "" {
		return fmt.Errorf("expected order with status, got empty")
	}
	if order.Customer.Name == "" {
		return fmt.Errorf("expected order with customer name, got empty")
	}
	return nil
}

func (g *GetOrderByIDSteps) theResponseShouldContainTheOrderItems() error {
	ctx := GetTestContext()
	order, ok := ctx.responseBody.(*responses.OrderResponse)
	if !ok {
		return fmt.Errorf("expected OrderResponse, got: %T", ctx.responseBody)
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("expected order with items, got empty")
	}
	return nil
}

// ===== Register Steps =====

func (g *GetOrderByIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for order operations$`, g.theUserIsAuthenticatedForOrderOperations)
	sc.Step(`^an order with ID (\d+) exists in shop (\d+)$`, g.anOrderWithIDExistsInShop)
	sc.Step(`^an order with ID (\d+) does not exist in shop (\d+)$`, g.anOrderWithIDDoesNotExistInShop)

	// When steps
	sc.Step(`^I send a get order by ID request for shop (\d+) and order (\d+)$`, g.iSendAGetOrderByIDRequestForShopAndOrder)
	sc.Step(`^I send a get order by ID request with invalid shop ID "([^"]*)" and order (\d+)$`, g.iSendAGetOrderByIDRequestWithInvalidShopIDAndOrder)

	// Then steps
	sc.Step(`^the response should contain the order details$`, g.theResponseShouldContainTheOrderDetails)
	sc.Step(`^the response should contain the order items$`, g.theResponseShouldContainTheOrderItems)
}

// Ensure sql package is used (for sql.NullString etc in other files)
var _ = sql.ErrNoRows
