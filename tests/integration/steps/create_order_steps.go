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
	scenarioCreateOrderValid          = "create-order-valid"
	scenarioCreateOrderDeliveryZone   = "create-order-delivery-zone"
	scenarioCreateOrderAddress        = "create-order-address"
	scenarioCreateOrderOptions        = "create-order-options"
	scenarioCreateOrderStoreNotFound  = "create-order-store-not-found"
	scenarioCreateOrderSubtotalWrong  = "create-order-subtotal-wrong"
	scenarioCreateOrderTotalWrong     = "create-order-total-wrong"
	scenarioCreateOrderHTTPValidation = "create-order-http-validation"
)

type CreateOrderSteps struct{}

func NewCreateOrderSteps() *CreateOrderSteps {
	return &CreateOrderSteps{}
}

// ===== Request map helpers (satisfy errcheck for type assertions) =====

func getOrderFromReq(req map[string]interface{}) (map[string]interface{}, error) {
	order, ok := req["order"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected order map in request")
	}
	return order, nil
}

func getSubMap(m map[string]interface{}, key string) (map[string]interface{}, error) {
	sub, ok := m[key].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map for key %q", key)
	}
	return sub, nil
}

func getItemsSlice(m map[string]interface{}) ([]map[string]interface{}, error) {
	items, ok := m["items"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected items slice in order")
	}
	return items, nil
}

// ===== Helper: build a valid order request body =====

func (g *CreateOrderSteps) buildValidOrderRequest() map[string]interface{} {
	return map[string]interface{}{
		"order": map[string]interface{}{
			"customer": map[string]interface{}{
				"name":  "John Doe",
				"phone": "1234567890",
				"email": "john@example.com",
			},
			"payment_method": map[string]interface{}{
				"id":   1,
				"name": "Transfer",
				"code": "transfer",
			},
			"delivery_method": map[string]interface{}{
				"id":            1,
				"name":          "Delivery",
				"code":          "delivery",
				"shipping_cost": 50.0,
			},
			"items": []map[string]interface{}{
				{
					"product": map[string]interface{}{
						"id":    10,
						"name":  "Test Product",
						"price": 100.0,
					},
					"quantity":   2,
					"unit_price": 100.0,
				},
			},
			"subtotal": 200.0,
			"total":    250.0,
		},
	}
}

// ===== Given Steps =====

func (g *CreateOrderSteps) aValidOrderRequestForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderValid
	ctx.requestBody = g.buildValidOrderRequest()
	return nil
}

func (g *CreateOrderSteps) aValidOrderRequestWithDeliveryZoneForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderDeliveryZone
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	dm, err := getSubMap(order, "delivery_method")
	if err != nil {
		return err
	}
	dm["delivery_zone"] = map[string]interface{}{
		"id":    1,
		"name":  "Zone A",
		"price": 50.0,
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) aValidOrderRequestWithCustomerAddressForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderAddress
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	customer, err := getSubMap(order, "customer")
	if err != nil {
		return err
	}
	customer["address"] = map[string]interface{}{
		"name":     "Main Street 123",
		"place_id": "ChIJ123",
		"lat":      -34.6037,
		"lng":      -58.3816,
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) aValidOrderRequestWithSelectedOptionsForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderOptions
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	items, err := getItemsSlice(order)
	if err != nil {
		return err
	}
	items[0]["selected_options"] = []map[string]interface{}{
		{
			"variant_id":   1,
			"variant_name": "Size",
			"option_id":    1,
			"option_name":  "Large",
			"option_price": 10.0,
			"quantity":     1,
		},
	}
	// Update prices to include option
	items[0]["unit_price"] = 110.0
	order["subtotal"] = 220.0
	order["total"] = 270.0
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithoutCustomerNameForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	customer, err := getSubMap(order, "customer")
	if err != nil {
		return err
	}
	customer["name"] = ""
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithoutPaymentMethodForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["payment_method"] = map[string]interface{}{
		"id":   0,
		"name": "Transfer",
		"code": "transfer",
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithoutItemsForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["items"] = []map[string]interface{}{}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithoutDeliveryMethodForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["delivery_method"] = map[string]interface{}{
		"id":            0,
		"name":          "Delivery",
		"code":          "delivery",
		"shipping_cost": 50.0,
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithoutDeliveryMethodNameForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["delivery_method"] = map[string]interface{}{
		"id":            1,
		"name":          "",
		"code":          "delivery",
		"shipping_cost": 50.0,
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithInvalidDeliveryZoneForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderHTTPValidation
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	dm, err := getSubMap(order, "delivery_method")
	if err != nil {
		return err
	}
	dm["delivery_zone"] = map[string]interface{}{
		"id":    0,
		"name":  "Zone A",
		"price": 50.0,
	}
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) aValidOrderRequestForANonexistentStore() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderStoreNotFound
	ctx.requestBody = g.buildValidOrderRequest()
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithWrongSubtotalForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderSubtotalWrong
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["subtotal"] = 999.0 // Wrong value
	order["total"] = 1049.0   // Keep total consistent with wrong subtotal
	ctx.requestBody = req
	return nil
}

func (g *CreateOrderSteps) anOrderRequestWithWrongTotalForStore(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCreateOrderTotalWrong
	req := g.buildValidOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["total"] = 999.0 // Wrong value
	ctx.requestBody = req
	return nil
}

// ===== When Steps =====

func (g *CreateOrderSteps) iSendACreateOrderRequestToStore(slug string) error {
	ctx := GetTestContext()

	if ctx.app == nil {
		if err := ctx.SetupCreateOrderTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupCreateOrderSQLExpectations(ctx, slug)

	jsonBody, err := json.Marshal(ctx.requestBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/stores/%s/orders", ctx.server.URL, slug)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseCreateOrderResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *CreateOrderSteps) setupCreateOrderSQLExpectations(ctx *TestContext, slug string) {
	// Shop query columns (13 columns from shopQueryBase)
	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram", "primary_color",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	// Product query columns (9 columns from GetByIDsAndShopID)
	productColumns := []string{
		"id", "name", "price", "is_stockeable", "stock", "is_active",
		"is_promotional", "promotional_price", "variants",
	}

	switch ctx.scenario {
	case scenarioCreateOrderValid, scenarioCreateOrderAddress:
		// Step 1: GetBySlug -> StoreService
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_config": {"fixed_price": 50}}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Step 2: GetByIDsAndShopID -> ValidateOrderItems
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: create_order stored procedure
		resultJSON := `{"id": 1, "order_number": "ORD-001", "status": "pending", "subtotal": 200, "shipping_cost": 50, "total": 250, "created_at": "2026-01-01T00:00:00Z"}`
		ctx.mockSQLMock.ExpectQuery("SELECT create_order").
			WillReturnRows(sqlmock.NewRows([]string{"create_order"}).AddRow(resultJSON))

	case scenarioCreateOrderDeliveryZone:
		// Step 1: GetBySlug with delivery zones
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_zones": [{"id": 1, "name": "Zone A", "price": 50}]}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Step 2: Products
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: create_order
		resultJSON := `{"id": 1, "order_number": "ORD-002", "status": "pending", "subtotal": 200, "shipping_cost": 50, "total": 250, "created_at": "2026-01-01T00:00:00Z"}`
		ctx.mockSQLMock.ExpectQuery("SELECT create_order").
			WillReturnRows(sqlmock.NewRows([]string{"create_order"}).AddRow(resultJSON))

	case scenarioCreateOrderOptions:
		// Step 1: GetBySlug
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_config": {"fixed_price": 50}}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Step 2: Products with variants
		variantsJSON := `[{"id": 1, "name": "Size", "order": 1, "selection_type": "single", "max_selections": 1, "is_required": false, "options": [{"id": 1, "name": "Large", "price": 10, "order": 1}]}]`
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, variantsJSON)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: create_order
		resultJSON := `{"id": 1, "order_number": "ORD-003", "status": "pending", "subtotal": 220, "shipping_cost": 50, "total": 270, "created_at": "2026-01-01T00:00:00Z"}`
		ctx.mockSQLMock.ExpectQuery("SELECT create_order").
			WillReturnRows(sqlmock.NewRows([]string{"create_order"}).AddRow(resultJSON))

	case scenarioCreateOrderStoreNotFound:
		// GetBySlug returns no rows
		emptyRows := sqlmock.NewRows(shopColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)

	case scenarioCreateOrderSubtotalWrong, scenarioCreateOrderTotalWrong:
		// Step 1: GetBySlug
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_config": {"fixed_price": 50}}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Step 2: Products (validation passes, fails at CalculateAndValidateTotals)
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// No step 3: fails before persist

	case scenarioCreateOrderHTTPValidation:
		// No SQL mock needed - fails at HTTP contract validation
	}
}

func (g *CreateOrderSteps) parseCreateOrderResponse(ctx *TestContext, resp *http.Response) {
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
		var response responses.CreateOrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response.Order
		}
	}
}

// ===== Then Steps =====

func (g *CreateOrderSteps) theResponseShouldContainTheCreatedOrder() error {
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
	if order.CreatedAt.IsZero() {
		return fmt.Errorf("expected order with created_at, got zero")
	}
	return nil
}

func (g *CreateOrderSteps) theCreatedOrderShouldHaveStatus(expectedStatus string) error {
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

func (g *CreateOrderSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a valid order request for store "([^"]*)"$`, g.aValidOrderRequestForStore)
	sc.Step(`^a valid order request with delivery zone for store "([^"]*)"$`, g.aValidOrderRequestWithDeliveryZoneForStore)
	sc.Step(`^a valid order request with customer address for store "([^"]*)"$`, g.aValidOrderRequestWithCustomerAddressForStore)
	sc.Step(`^a valid order request with selected options for store "([^"]*)"$`, g.aValidOrderRequestWithSelectedOptionsForStore)
	sc.Step(`^an order request without customer name for store "([^"]*)"$`, g.anOrderRequestWithoutCustomerNameForStore)
	sc.Step(`^an order request without payment method for store "([^"]*)"$`, g.anOrderRequestWithoutPaymentMethodForStore)
	sc.Step(`^an order request without items for store "([^"]*)"$`, g.anOrderRequestWithoutItemsForStore)
	sc.Step(`^an order request without delivery method for store "([^"]*)"$`, g.anOrderRequestWithoutDeliveryMethodForStore)
	sc.Step(`^an order request without delivery method name for store "([^"]*)"$`, g.anOrderRequestWithoutDeliveryMethodNameForStore)
	sc.Step(`^an order request with invalid delivery zone for store "([^"]*)"$`, g.anOrderRequestWithInvalidDeliveryZoneForStore)
	sc.Step(`^a valid order request for a nonexistent store$`, g.aValidOrderRequestForANonexistentStore)
	sc.Step(`^an order request with wrong subtotal for store "([^"]*)"$`, g.anOrderRequestWithWrongSubtotalForStore)
	sc.Step(`^an order request with wrong total for store "([^"]*)"$`, g.anOrderRequestWithWrongTotalForStore)

	// When steps
	sc.Step(`^I send a create order request to store "([^"]*)"$`, g.iSendACreateOrderRequestToStore)

	// Then steps
	sc.Step(`^the response should contain the created order$`, g.theResponseShouldContainTheCreatedOrder)
	sc.Step(`^the created order should have status "([^"]*)"$`, g.theCreatedOrderShouldHaveStatus)
}

// Ensure time is used
var _ = time.Now
