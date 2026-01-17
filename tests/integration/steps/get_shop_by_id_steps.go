package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

const (
	scenarioShopExists         = "shop-exists"
	scenarioShopExistsNoImages = "shop-exists-no-images"
	scenarioShopNotFound       = "shop-not-found"
	scenarioShopNotOwned       = "shop-not-owned"

	// Test JSON data constants
	testAddressJSON            = `{"id": 1, "name": "Main Street 123", "place_id": "ChIJ123", "lat": -34.6037, "lng": -58.3816}`
	testPaymentMethodsJSON     = `[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`
	testDeliveryMethodsJSON    = `[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true}]`
	testOperatingSchedulesJSON = `[{"id": 1, "day_of_week": 1, "open_time": "09:00", "close_time": "18:00"}]`
)

type GetShopByIDSteps struct{}

func NewGetShopByIDSteps() *GetShopByIDSteps {
	return &GetShopByIDSteps{}
}

// ===== Given Steps =====

func (g *GetShopByIDSteps) theUserIsAuthenticatedForShopOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GetShopByIDSteps) theUserIsNotAuthenticatedForShop() error {
	ctx := GetTestContext()
	ctx.authToken = "" // Clear auth token
	return nil
}

func (g *GetShopByIDSteps) aShopWithIDExists(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetShopByIDSteps) aShopWithIDExistsWithoutImages(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopExistsNoImages
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetShopByIDSteps) aShopWithIDDoesNotExist(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopNotFound
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetShopByIDSteps) aShopWithIDExistsButUserDoesNotOwnIt(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopNotOwned
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

// ===== When Steps =====

func (g *GetShopByIDSteps) iSendAGetShopByIDRequestForShop(shopID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetShopByIDSQLExpectations(shopID)

	// Build URL and make request
	url := fmt.Sprintf("%s/shops/%d", ctx.server.URL, shopID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add authorization header
	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseShopResponse(ctx, resp)

	return nil
}

func (g *GetShopByIDSteps) iSendAGetShopByIDRequestWithInvalidID(invalidID string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}

	// Build URL with invalid ID
	url := fmt.Sprintf("%s/shops/%s", ctx.server.URL, invalidID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add authorization header
	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseShopResponse(ctx, resp)

	return nil
}

func (g *GetShopByIDSteps) iSendAnUnauthenticatedGetShopByIDRequestForShop(shopID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}

	// Build URL and make request WITHOUT auth header
	url := fmt.Sprintf("%s/shops/%d", ctx.server.URL, shopID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// No authorization header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseShopResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *GetShopByIDSteps) setupGetShopByIDSQLExpectations(shopID int) {
	ctx := GetTestContext()

	// Columns returned by GetByID query (complex query with multiple LEFT JOIN LATERAL)
	columns := []string{
		"id", "name", "slug", "email", "phone", "instagram", "created_at",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules",
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioShopExists:
		// Mock shop with all relations
		imagesJSON := `[{"id": 1, "url": "https://cloudinary.com/logo.jpg", "type": "logo"}, {"id": 2, "url": "https://cloudinary.com/cover.jpg", "type": "cover"}]`

		rows := sqlmock.NewRows(columns).
			AddRow(shopID, "Test Shop", "test-shop", "test@shop.com", "+54111234567", "@testshop", now,
				imagesJSON, testAddressJSON, testPaymentMethodsJSON, testDeliveryMethodsJSON, testOperatingSchedulesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(shopID).
			WillReturnRows(rows)

	case scenarioShopExistsNoImages:
		// Mock shop without images
		rows := sqlmock.NewRows(columns).
			AddRow(shopID, "Test Shop No Images", "test-shop-no-images", "test@shop.com", "+54111234567", "@testshop", now,
				"[]", testAddressJSON, testPaymentMethodsJSON, testDeliveryMethodsJSON, testEmptySchedulesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(shopID).
			WillReturnRows(rows)

	case scenarioShopNotFound:
		// Mock no rows found
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(shopID).
			WillReturnRows(emptyRows)

	case scenarioShopNotOwned:
		// This scenario doesn't require SQL mock - it's handled by authorization check in handler
		// The handler checks if user owns the shop before making DB call
	}
}

func (g *GetShopByIDSteps) parseShopResponse(ctx *TestContext, resp *http.Response) {
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
		var shop models.Shop
		if err := json.NewDecoder(resp.Body).Decode(&shop); err == nil {
			ctx.responseBody = shop
		}
	}
}

// ===== Then Steps =====

func (g *GetShopByIDSteps) theResponseShouldContainTheShopDetails() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if shop.ID == 0 {
		return fmt.Errorf("expected shop with ID, got ID 0")
	}
	if shop.Name == "" {
		return fmt.Errorf("expected shop with name, got empty name")
	}
	if shop.Slug == "" {
		return fmt.Errorf("expected shop with slug, got empty slug")
	}
	return nil
}

func (g *GetShopByIDSteps) theResponseShouldContainTheShopImages() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if len(shop.Images) == 0 {
		return fmt.Errorf("expected shop with images, got nil or empty images")
	}
	return nil
}

func (g *GetShopByIDSteps) theResponseShouldNotContainShopImages() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if len(shop.Images) > 0 {
		return fmt.Errorf("expected shop without images, got %d images", len(shop.Images))
	}
	return nil
}

func (g *GetShopByIDSteps) theResponseShouldContainTheShopAddress() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if shop.Address == nil {
		return fmt.Errorf("expected shop with address, got nil address")
	}
	return nil
}

func (g *GetShopByIDSteps) theResponseShouldContainTheShopPaymentMethods() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if len(shop.PaymentMethods) == 0 {
		return fmt.Errorf("expected shop with payment methods, got nil or empty")
	}
	return nil
}

func (g *GetShopByIDSteps) theResponseShouldContainTheShopDeliveryMethods() error {
	ctx := GetTestContext()
	shop, ok := ctx.responseBody.(models.Shop)
	if !ok {
		return fmt.Errorf("expected Shop, got: %T", ctx.responseBody)
	}
	if len(shop.DeliveryMethods) == 0 {
		return fmt.Errorf("expected shop with delivery methods, got nil or empty")
	}
	return nil
}

// ===== Register Steps =====

func (g *GetShopByIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for shop operations$`, g.theUserIsAuthenticatedForShopOperations)
	sc.Step(`^the user is not authenticated$`, g.theUserIsNotAuthenticatedForShop)
	sc.Step(`^a shop with ID (\d+) exists$`, g.aShopWithIDExists)
	sc.Step(`^a shop with ID (\d+) exists without images$`, g.aShopWithIDExistsWithoutImages)
	sc.Step(`^a shop with ID (\d+) does not exist$`, g.aShopWithIDDoesNotExist)
	sc.Step(`^a shop with ID (\d+) exists but user does not own it$`, g.aShopWithIDExistsButUserDoesNotOwnIt)

	// When steps
	sc.Step(`^I send a get shop by ID request for shop (\d+)$`, g.iSendAGetShopByIDRequestForShop)
	sc.Step(`^I send a get shop by ID request with invalid ID "([^"]*)"$`, g.iSendAGetShopByIDRequestWithInvalidID)
	sc.Step(`^I send an unauthenticated get shop by ID request for shop (\d+)$`, g.iSendAnUnauthenticatedGetShopByIDRequestForShop)

	// Then steps
	sc.Step(`^the response should contain the shop details$`, g.theResponseShouldContainTheShopDetails)
	sc.Step(`^the response should contain the shop images$`, g.theResponseShouldContainTheShopImages)
	sc.Step(`^the response should not contain shop images$`, g.theResponseShouldNotContainShopImages)
	sc.Step(`^the response should contain the shop address$`, g.theResponseShouldContainTheShopAddress)
	sc.Step(`^the response should contain the shop payment methods$`, g.theResponseShouldContainTheShopPaymentMethods)
	sc.Step(`^the response should contain the shop delivery methods$`, g.theResponseShouldContainTheShopDeliveryMethods)
}
