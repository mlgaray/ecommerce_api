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
	scenarioUpdateOrderEditable      = "update-order-editable"
	scenarioUpdateOrderDeliveryZone  = "update-order-delivery-zone"
	scenarioUpdateOrderCompleted     = "update-order-completed"
	scenarioUpdateOrderCancelled     = "update-order-cancelled"
	scenarioUpdateOrderNotFound      = "update-order-not-found"
	scenarioUpdateOrderSubtotalWrong = "update-order-subtotal-wrong"

	updateOrderResultJSON       = `{"updated_at": "2026-01-02T00:00:00Z", "subtotal": 300, "total": 350}`
	updateOrderUpdatedItemsJSON = `[{"id":1,"product_id":10,"product_name":"Test Product","product_image_url":"","base_price":100,"is_promotional":false,"promotional_price":0,"quantity":3,"unit_price":100,"total_price":300,"notes":"","selected_options":[]}]`
)

type UpdateOrderSteps struct{}

func NewUpdateOrderSteps() *UpdateOrderSteps {
	return &UpdateOrderSteps{}
}

// ===== Helper: build update request body =====

func (g *UpdateOrderSteps) buildUpdateOrderRequest() map[string]interface{} {
	return map[string]interface{}{
		"order": map[string]interface{}{
			"customer": map[string]interface{}{
				"name":  "Jane Doe Updated",
				"phone": "9876543210",
				"email": "jane@example.com",
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
					"quantity":   3,
					"unit_price": 100.0,
				},
			},
			"subtotal": 300.0,
			"total":    350.0,
		},
	}
}

// ===== Given Steps =====

func (g *UpdateOrderSteps) anEditableOrderExistsInShop(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderEditable
	return nil
}

func (g *UpdateOrderSteps) anEditableOrderExistsInShopWithDeliveryZones(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderDeliveryZone
	return nil
}

func (g *UpdateOrderSteps) aCompletedOrderExistsInShop(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderCompleted
	return nil
}

func (g *UpdateOrderSteps) aCancelledOrderExistsInShop(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderCancelled
	return nil
}

func (g *UpdateOrderSteps) anOrderDoesNotExistInShopForUpdate(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderNotFound
	return nil
}

func (g *UpdateOrderSteps) anEditableOrderExistsInShopForSubtotalValidation(orderID int, shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateOrderSubtotalWrong
	return nil
}

// ===== When Steps =====

func (g *UpdateOrderSteps) iSendAnUpdateOrderRequest(shopID int, orderID int) error {
	ctx := GetTestContext()
	ctx.requestBody = g.buildUpdateOrderRequest()
	return g.sendUpdateRequest(ctx, shopID, orderID)
}

func (g *UpdateOrderSteps) iSendAnUpdateOrderRequestWithDeliveryZone(shopID int, orderID int) error {
	ctx := GetTestContext()
	req := g.buildUpdateOrderRequest()
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
	return g.sendUpdateRequest(ctx, shopID, orderID)
}

func (g *UpdateOrderSteps) iSendAnUpdateOrderRequestWithWrongSubtotal(shopID int, orderID int) error {
	ctx := GetTestContext()
	req := g.buildUpdateOrderRequest()
	order, err := getOrderFromReq(req)
	if err != nil {
		return err
	}
	order["subtotal"] = 999.0
	order["total"] = 1049.0
	ctx.requestBody = req
	return g.sendUpdateRequest(ctx, shopID, orderID)
}

func (g *UpdateOrderSteps) sendUpdateRequest(ctx *TestContext, shopID int, orderID int) error {
	if ctx.app == nil {
		if err := ctx.SetupOrderTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations
	g.setupUpdateOrderSQLExpectations(ctx, shopID, orderID)

	jsonBody, err := json.Marshal(ctx.requestBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/shops/%d/orders/%d", ctx.server.URL, shopID, orderID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonBody))
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
	g.parseUpdateOrderResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *UpdateOrderSteps) setupUpdateOrderSQLExpectations(ctx *TestContext, shopID int, orderID int) {
	getByIDColumns := []string{
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

	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram", "primary_color",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	productColumns := []string{
		"id", "name", "price", "is_stockeable", "stock", "is_active",
		"is_promotional", "promotional_price", "variants",
	}

	now := time.Now()
	itemsJSON := `[{"id":1,"product_id":10,"product_name":"Test Product","product_image_url":"","base_price":100,"is_promotional":false,"promotional_price":0,"quantity":2,"unit_price":100,"total_price":200,"notes":"","selected_options":[]}]`

	switch ctx.scenario {
	case scenarioUpdateOrderEditable:
		// Step 1: GetByID - existing editable order (pending)
		rows1 := sqlmock.NewRows(getByIDColumns).
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
			WillReturnRows(rows1)

		// Step 2: ValidateOrderItems -> GetByIDsAndShopID
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: ValidateDeliveryMethod -> GetBySlug
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(shopID, "Test Store", "test-store", "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_config": {"fixed_price": 50}}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WillReturnRows(shopRows)

		// Step 4: update_order stored procedure
		ctx.mockSQLMock.ExpectQuery("SELECT update_order").
			WillReturnRows(sqlmock.NewRows([]string{"update_order"}).AddRow(updateOrderResultJSON))

		// Step 5: GetByID re-fetch
		rows2 := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", "pending",
				shopID, "Test Store", "test-store",
				"Jane Doe Updated", "9876543210", "jane@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				nil, nil, nil,
				300.0, 50.0, 0.0, 350.0,
				nil, nil, nil, nil, nil,
				now, now,
				updateOrderUpdatedItemsJSON,
			)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows2)

	case scenarioUpdateOrderDeliveryZone:
		// Step 1: GetByID
		rows1 := sqlmock.NewRows(getByIDColumns).
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
			WillReturnRows(rows1)

		// Step 2: Products
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: GetBySlug (with delivery zones)
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(shopID, "Test Store", "test-store", "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_zones": [{"id": 1, "name": "Zone A", "price": 50}]}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WillReturnRows(shopRows)

		// Step 4: update_order
		ctx.mockSQLMock.ExpectQuery("SELECT update_order").
			WillReturnRows(sqlmock.NewRows([]string{"update_order"}).AddRow(updateOrderResultJSON))

		// Step 5: GetByID re-fetch
		rows2 := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", "pending",
				shopID, "Test Store", "test-store",
				"Jane Doe Updated", "9876543210", "jane@example.com",
				nil, nil, nil, nil,
				1, "transfer", "Transfer",
				1, "delivery", "Delivery",
				1, "Zone A", 50.0,
				300.0, 50.0, 0.0, 350.0,
				nil, nil, nil, nil, nil,
				now, now,
				updateOrderUpdatedItemsJSON,
			)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(rows2)

	case scenarioUpdateOrderCompleted:
		// GetByID returns completed order
		rows := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", "completed",
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

	case scenarioUpdateOrderCancelled:
		// GetByID returns cancelled order
		rows := sqlmock.NewRows(getByIDColumns).
			AddRow(
				orderID, "ORD-001", "cancelled",
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

	case scenarioUpdateOrderNotFound:
		emptyRows := sqlmock.NewRows(getByIDColumns)
		ctx.mockSQLMock.ExpectQuery("WITH items_with_selections").
			WithArgs(orderID, shopID).
			WillReturnRows(emptyRows)

	case scenarioUpdateOrderSubtotalWrong:
		// Step 1: GetByID
		rows1 := sqlmock.NewRows(getByIDColumns).
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
			WillReturnRows(rows1)

		// Step 2: Products
		productRows := sqlmock.NewRows(productColumns).
			AddRow(10, "Test Product", 100.0, false, 0, true, false, 0, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

		// Step 3: GetBySlug for delivery validation
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(shopID, "Test Store", "test-store", "store@test.com", "+54111234567", "@teststore", nil,
				"[]",
				`{"id": 1, "name": "Main St", "place_id": "ChIJ", "lat": -34.6, "lng": -58.38}`,
				`[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`,
				`[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true, "delivery_config": {"fixed_price": 50}}]`,
				"[]", nil)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WillReturnRows(shopRows)

		// No step 4: fails at CalculateAndValidateTotals
	}
}

func (g *UpdateOrderSteps) parseUpdateOrderResponse(ctx *TestContext, resp *http.Response) {
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
		var response responses.UpdateOrderResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response.Order
		}
	}
}

// ===== Then Steps =====

func (g *UpdateOrderSteps) theResponseShouldContainTheUpdatedOrder() error {
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
	return nil
}

// ===== Register Steps =====

func (g *UpdateOrderSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^an editable order with ID (\d+) exists in shop (\d+)$`, g.anEditableOrderExistsInShop)
	sc.Step(`^an editable order with ID (\d+) exists in shop (\d+) with delivery zones$`, g.anEditableOrderExistsInShopWithDeliveryZones)
	sc.Step(`^a completed order with ID (\d+) exists in shop (\d+)$`, g.aCompletedOrderExistsInShop)
	sc.Step(`^a cancelled order with ID (\d+) exists in shop (\d+)$`, g.aCancelledOrderExistsInShop)
	sc.Step(`^an order with ID (\d+) does not exist in shop (\d+) for update$`, g.anOrderDoesNotExistInShopForUpdate)
	sc.Step(`^an editable order with ID (\d+) exists in shop (\d+) for subtotal validation$`, g.anEditableOrderExistsInShopForSubtotalValidation)

	// When steps
	sc.Step(`^I send an update order request for shop (\d+) and order (\d+)$`, g.iSendAnUpdateOrderRequest)
	sc.Step(`^I send an update order request with delivery zone for shop (\d+) and order (\d+)$`, g.iSendAnUpdateOrderRequestWithDeliveryZone)
	sc.Step(`^I send an update order request with wrong subtotal for shop (\d+) and order (\d+)$`, g.iSendAnUpdateOrderRequestWithWrongSubtotal)

	// Then steps
	sc.Step(`^the response should contain the updated order$`, g.theResponseShouldContainTheUpdatedOrder)
}
