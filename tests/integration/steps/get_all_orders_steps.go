package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	scenarioOrdersExist = "orders-exist"
	scenarioNoOrders    = "no-orders"
)

type GetAllOrdersSteps struct{}

func NewGetAllOrdersSteps() *GetAllOrdersSteps {
	return &GetAllOrdersSteps{}
}

// ===== Given Steps =====

func (g *GetAllOrdersSteps) ordersExistInShop(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioOrdersExist
	_ = shopID
	return nil
}

func (g *GetAllOrdersSteps) noOrdersExistInShop(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioNoOrders
	_ = shopID
	return nil
}

// ===== When Steps =====

func (g *GetAllOrdersSteps) iSendAGetAllOrdersRequestForShop(shopID int) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetAllSQLExpectations(ctx, shopID)

	url := fmt.Sprintf("%s/shops/%d/orders", ctx.server.URL, shopID)
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
	g.parseGetAllResponse(ctx, resp)

	return nil
}

func (g *GetAllOrdersSteps) iSendAGetAllOrdersRequestWithInvalidShopID(invalidShopID string) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/shops/%s/orders", ctx.server.URL, invalidShopID)
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
	g.parseGetAllResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *GetAllOrdersSteps) setupGetAllSQLExpectations(ctx *TestContext, shopID int) {
	// 26 columns returned by GetAllByShopIDWithFilters query (lightweight - no items)
	columns := []string{
		"id", "order_number", "status",
		"customer_name", "customer_phone", "customer_email", "customer_address_name",
		"payment_method_id", "payment_method_code", "payment_method_name",
		"delivery_method_id", "delivery_method_code", "delivery_method_name",
		"delivery_zone_name",
		"subtotal", "shipping_cost", "discount", "total",
		"coupon_code", "coupon_type", "coupon_value",
		"coupon_discount_amount", "coupon_min_order_amount",
		"created_at", "updated_at",
		"items_count",
	}

	// Count query columns (for first page total count)
	countColumns := []string{"count"}

	switch ctx.scenario {
	case scenarioOrdersExist:
		now := time.Now()
		rows := sqlmock.NewRows(columns).
			AddRow(
				1, "ORD-001", "pending",
				"John Doe", "1234567890", "john@example.com", nil,
				1, "transfer", "Transferencia",
				1, "delivery", "Delivery",
				nil,
				200.0, 50.0, 0.0, 250.0,
				nil, nil, nil, nil, nil,
				now, now,
				2,
			).
			AddRow(
				2, "ORD-002", "confirmed",
				"Jane Doe", "0987654321", "jane@example.com", nil,
				1, "cash", "Efectivo",
				1, "pickup", "Retiro",
				nil,
				100.0, 0.0, 0.0, 100.0,
				nil, nil, nil, nil, nil,
				now, now,
				1,
			)

		ctx.mockSQLMock.ExpectQuery("SELECT").
			WithArgs(shopID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		countRows := sqlmock.NewRows(countColumns).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WithArgs(shopID).
			WillReturnRows(countRows)

	case scenarioNoOrders:
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT").
			WithArgs(shopID, sqlmock.AnyArg()).
			WillReturnRows(emptyRows)

		countRows := sqlmock.NewRows(countColumns).AddRow(0)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WithArgs(shopID).
			WillReturnRows(countRows)
	}
}

func (g *GetAllOrdersSteps) parseGetAllResponse(ctx *TestContext, resp *http.Response) {
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
		var response responses.ListOrdersResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = &response
		}
	}
}

// ===== Then Steps =====

func (g *GetAllOrdersSteps) theResponseShouldContainAListOfOrders() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(*responses.ListOrdersResponse)
	if !ok {
		return fmt.Errorf("expected ListOrdersResponse, got: %T", ctx.responseBody)
	}
	if len(response.Orders) == 0 {
		return fmt.Errorf("expected orders in response, got empty list")
	}
	return nil
}

func (g *GetAllOrdersSteps) theResponseShouldContainAnEmptyListOfOrders() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(*responses.ListOrdersResponse)
	if !ok {
		return fmt.Errorf("expected ListOrdersResponse, got: %T", ctx.responseBody)
	}
	if len(response.Orders) != 0 {
		return fmt.Errorf("expected empty orders list, got %d orders", len(response.Orders))
	}
	return nil
}

// ===== Register Steps =====

func (g *GetAllOrdersSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^orders exist in shop (\d+)$`, g.ordersExistInShop)
	sc.Step(`^no orders exist in shop (\d+)$`, g.noOrdersExistInShop)

	// When steps
	sc.Step(`^I send a get all orders request for shop (\d+)$`, g.iSendAGetAllOrdersRequestForShop)
	sc.Step(`^I send a get all orders request with invalid shop ID "([^"]*)"$`, g.iSendAGetAllOrdersRequestWithInvalidShopID)

	// Then steps
	sc.Step(`^the response should contain a list of orders$`, g.theResponseShouldContainAListOfOrders)
	sc.Step(`^the response should contain an empty list of orders$`, g.theResponseShouldContainAnEmptyListOfOrders)
}
