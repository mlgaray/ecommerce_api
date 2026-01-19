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
	scenarioStoreProductByIDExists   = "store-product-by-id-exists"
	scenarioStoreProductByIDNotFound = "store-product-by-id-not-found"
	scenarioStoreProductByIDInactive = "store-product-by-id-inactive"
)

type GetStoreProductByIDSteps struct{}

func NewGetStoreProductByIDSteps() *GetStoreProductByIDSteps {
	return &GetStoreProductByIDSteps{}
}

// ===== Given Steps =====

func (g *GetStoreProductByIDSteps) aStoreWithSlugExistsWithAnInactiveProductWithID(slug string, productID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductByIDInactive
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	return nil
}

// ===== When Steps =====

func (g *GetStoreProductByIDSteps) iSendAGetStoreProductByIDRequestFor(slug string, productID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// If requesting a different slug than what was set up, treat as "not found"
	if ctx.pathParams != nil && ctx.pathParams["slug"] != "" && ctx.pathParams["slug"] != slug {
		ctx.scenario = scenarioStoreNotFound
	}

	// Set scenario based on product ID if not already set
	if ctx.scenario == "" {
		ctx.scenario = scenarioStoreProductByIDExists
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductByIDSQLExpectations(slug, productID)

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s/products/%d", ctx.server.URL, slug, productID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseStoreProductByIDResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductByIDSteps) iSendAGetStoreProductByIDRequestForWithInvalidProductID(slug, invalidProductID string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// No SQL mock setup needed - handler will reject invalid product ID before DB query

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s/products/%s", ctx.server.URL, slug, invalidProductID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseStoreProductByIDResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductByIDSteps) iSendAnUnauthenticatedGetStoreProductByIDRequestFor(slug string, productID int) error {
	// Store endpoints are always unauthenticated, so this is the same as the regular request
	return g.iSendAGetStoreProductByIDRequestFor(slug, productID)
}

// ===== SQL Mock Setup =====

func (g *GetStoreProductByIDSteps) setupGetStoreProductByIDSQLExpectations(slug string, productID int) {
	ctx := GetTestContext()

	// Shop columns for GetBySlug query - must match shop_repository.go scan order
	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	// Product columns for GetByIDAndShopID query - must match product_repository.go scan order:
	// ID, Name, Description, Price, Stock, MinimumStock, IsActive, IsHighlighted, IsPromotional, PromotionalPrice, CreatedAt, Category.ID, Category.Name, Category.Description, imagesJSON, variantsJSON
	productColumns := []string{
		"id", "name", "description", "price", "stock", "minimum_stock",
		"is_active", "is_highlighted", "is_promotional", "promotional_price",
		"created_at", "category_id", "category_name", "category_description", "images", "variants",
	}

	now := time.Now()
	productImagesJSON := testProductImagesJSON
	productVariantsJSON := `[{"id":1,"name":"Size","order":1,"selection_type":"single","max_selections":1,"is_required":true,"options":[{"id":1,"name":"Small","price":0,"order":1},{"id":2,"name":"Large","price":2,"order":2}]}]`

	switch ctx.scenario {
	case scenarioStoreProductByIDExists, scenarioStoreProductsExists:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock product by ID - active product
		// Note: Query uses CTEs, so we use (?s) to match across newlines
		productRows := sqlmock.NewRows(productColumns).
			AddRow(productID, "Test Product", "Test product description", 99.99, 10, 2, true, false, false, 0.0, now, 1, "Electronics", "Electronics desc", productImagesJSON, productVariantsJSON)

		ctx.mockSQLMock.ExpectQuery("(?s)FROM products p").
			WillReturnRows(productRows)

	case scenarioStoreProductByIDNotFound:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock product not found - empty result
		// Note: Query uses CTEs, so we use (?s) to match across newlines
		productRows := sqlmock.NewRows(productColumns)
		ctx.mockSQLMock.ExpectQuery("(?s)FROM products p").
			WillReturnRows(productRows)

	case scenarioStoreProductByIDInactive:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock inactive product
		// Note: Query uses CTEs, so we use (?s) to match across newlines
		productRows := sqlmock.NewRows(productColumns).
			AddRow(productID, "Inactive Product", "Inactive product description", 99.99, 10, 2, false, false, false, 0.0, now, 1, "Electronics", "Electronics desc", productImagesJSON, productVariantsJSON)

		ctx.mockSQLMock.ExpectQuery("(?s)FROM products p").
			WillReturnRows(productRows)

	case scenarioStoreNotFound:
		// Mock shop not found
		emptyRows := sqlmock.NewRows(shopColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)
	}
}

func (g *GetStoreProductByIDSteps) parseStoreProductByIDResponse(ctx *TestContext, resp *http.Response) {
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
		var response models.Product
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response
		}
	}
}

// ===== Then Steps =====

func (g *GetStoreProductByIDSteps) theResponseShouldContainProductDetails() error {
	ctx := GetTestContext()
	product, ok := ctx.responseBody.(models.Product)
	if !ok {
		return fmt.Errorf("expected Product, got: %T", ctx.responseBody)
	}
	if product.ID == 0 {
		return fmt.Errorf("expected product ID, got 0")
	}
	if product.Name == "" {
		return fmt.Errorf("expected product name, got empty")
	}
	return nil
}

func (g *GetStoreProductByIDSteps) theResponseShouldIncludeProductVariantsAndOptions() error {
	ctx := GetTestContext()
	product, ok := ctx.responseBody.(models.Product)
	if !ok {
		return fmt.Errorf("expected Product, got: %T", ctx.responseBody)
	}
	if product.Variants == nil {
		return fmt.Errorf("expected variants, got nil")
	}
	if len(product.Variants) == 0 {
		return fmt.Errorf("expected variants in response, got empty")
	}
	if len(product.Variants[0].Options) == 0 {
		return fmt.Errorf("expected options in variants, got empty")
	}
	return nil
}

// ===== Register Steps =====

func (g *GetStoreProductByIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a store with slug "([^"]*)" exists with an inactive product with id (\d+)$`, g.aStoreWithSlugExistsWithAnInactiveProductWithID)

	// When steps
	sc.Step(`^I send a get store product by id request for "([^"]*)" with product_id (\d+)$`, g.iSendAGetStoreProductByIDRequestFor)
	sc.Step(`^I send a get store product by id request for "([^"]*)" with invalid product_id "([^"]*)"$`, g.iSendAGetStoreProductByIDRequestForWithInvalidProductID)
	sc.Step(`^I send an unauthenticated get store product by id request for "([^"]*)" with product_id (\d+)$`, g.iSendAnUnauthenticatedGetStoreProductByIDRequestFor)
	sc.Step(`^I send a get store product by id request for "([^"]*)" with product_id (-?\d+)$`, g.iSendAGetStoreProductByIDRequestFor)

	// Then steps
	sc.Step(`^the response should contain product details$`, g.theResponseShouldContainProductDetails)
	sc.Step(`^the response should include product variants and options$`, g.theResponseShouldIncludeProductVariantsAndOptions)
}
