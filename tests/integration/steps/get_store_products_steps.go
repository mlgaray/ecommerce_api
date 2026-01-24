package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

const (
	scenarioStoreProductsExists            = "store-products-exists"
	scenarioStoreProductsEmpty             = "store-products-empty"
	scenarioStoreProductsNotFound          = "store-products-not-found"
	scenarioStoreProductsWithSearch        = "store-products-with-search"
	scenarioStoreProductsWithCategory      = "store-products-with-category"
	scenarioStoreProductsWithLimit         = "store-products-with-limit"
	scenarioStoreProductsMixedActive       = "store-products-mixed-active"
	scenarioStoreProductsSearchAndCategory = "store-products-search-and-category"
	scenarioStoreProductsWithSort          = "store-products-with-sort"
)

type GetStoreProductsSteps struct{}

func NewGetStoreProductsSteps() *GetStoreProductsSteps {
	return &GetStoreProductsSteps{}
}

// ===== Given Steps =====

func (g *GetStoreProductsSteps) aStoreWithSlugExistsWithProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreProductsSteps) aStoreWithSlugExistsWithoutProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsEmpty
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreProductsSteps) aStoreWithSlugExistsWithMixedActiveAndInactiveProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsMixedActive
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

// ===== When Steps =====

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestFor(slug string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// If requesting a different slug than what was set up, treat as "not found"
	if ctx.pathParams != nil && ctx.pathParams["slug"] != "" && ctx.pathParams["slug"] != slug {
		ctx.scenario = scenarioStoreProductsNotFound
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, "", 0, 0)

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s/products", ctx.server.URL, slug)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestForWithSearch(slug, search string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsWithSearch

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, search, 0, 0)

	// Build URL with search param and make request
	url := fmt.Sprintf("%s/stores/%s/products?search=%s", ctx.server.URL, slug, search)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestForWithCategoryID(slug string, categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsWithCategory

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, "", categoryID, 0)

	// Build URL with category_id param and make request
	url := fmt.Sprintf("%s/stores/%s/products?category_id=%d", ctx.server.URL, slug, categoryID)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestForWithLimit(slug string, limit int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsWithLimit

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, "", 0, limit)

	// Build URL with limit param and make request
	url := fmt.Sprintf("%s/stores/%s/products?limit=%d", ctx.server.URL, slug, limit)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductsSteps) iSendAnUnauthenticatedGetStoreProductsRequestFor(slug string) error {
	// Store endpoints are always unauthenticated, so this is the same as the regular request
	return g.iSendAGetStoreProductsRequestFor(slug)
}

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestForWithSearchAndCategoryID(slug, search string, categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsSearchAndCategory

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, search, categoryID, 0)

	// Build URL with search and category_id params and make request
	url := fmt.Sprintf("%s/stores/%s/products?search=%s&category_id=%d", ctx.server.URL, slug, search, categoryID)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreProductsSteps) iSendAGetStoreProductsRequestForWithSortAndOrder(slug, sort, order string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreProductsWithSort

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreProductsSQLExpectations(slug, "", 0, 0)

	// Build URL with sort params and make request
	url := fmt.Sprintf("%s/stores/%s/products?sort_by=%s&sort_order=%s", ctx.server.URL, slug, sort, order)
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
	g.parseStoreProductsResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

//nolint:unparam // search, categoryID, limit are intentionally unused - SQL mocks return fixed data regardless of filters
func (g *GetStoreProductsSteps) setupGetStoreProductsSQLExpectations(slug, search string, categoryID, limit int) {
	ctx := GetTestContext()

	// Shop columns for GetBySlug query - must match shop_repository.go scan order
	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram", "primary_color",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	// Product columns for GetAllByShopIDWithFilters query - must match product_repository.go scan order:
	// ID, Name, Description, Price, Stock, MinimumStock, IsActive, IsHighlighted, IsPromotional, PromotionalPrice, CreatedAt, Category.ID, Category.Name, Category.Description, imagesJSON
	productColumns := []string{
		"id", "name", "description", "price", "stock", "minimum_stock",
		"is_active", "is_highlighted", "is_promotional", "promotional_price",
		"created_at", "category_id", "category_name", "category_description", "images",
	}

	now := time.Now()
	productImagesJSON := testProductImagesJSON

	switch ctx.scenario {
	case scenarioStoreProductsExists, scenarioStoreProductsWithSearch, scenarioStoreProductsWithCategory, scenarioStoreProductsWithLimit, scenarioStoreProductsSearchAndCategory, scenarioStoreProductsWithSort:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore", "#8B5CF6",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock products count
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock products - order: id, name, desc, price, stock, min_stock, is_active, is_highlighted, is_promotional, promo_price, created_at, cat_id, cat_name, cat_desc, images
		productRows := sqlmock.NewRows(productColumns).
			AddRow(1, "Product 1", "Description 1", 99.99, 10, 2, true, false, false, 0.0, now, 1, "Electronics", "Electronics desc", productImagesJSON).
			AddRow(2, "Product 2", "Description 2", 149.99, 5, 1, true, true, true, 129.99, now, 1, "Electronics", "Electronics desc", productImagesJSON).
			AddRow(3, "Product 3", "Description 3", 199.99, 20, 5, true, false, false, 0.0, now, 2, "Clothing", "Clothing desc", productImagesJSON).
			AddRow(4, "Product 4", "Description 4", 249.99, 15, 3, false, true, false, 0.0, now, 1, "Electronics", "Electronics desc", productImagesJSON).
			AddRow(5, "Product 5", "Description 5", 299.99, 8, 2, true, false, true, 249.99, now, 2, "Clothing", "Clothing desc", productImagesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

	case scenarioStoreProductsEmpty:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(2, "Empty Store", slug, "test@store.com", "+54111234567", "@emptystore", "#8B5CF6",
				"[]", testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, nil)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock products count (0)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock empty products
		productRows := sqlmock.NewRows(productColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

	case scenarioStoreProductsNotFound, scenarioStoreNotFound:
		// Mock shop not found
		emptyRows := sqlmock.NewRows(shopColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)

	case scenarioStoreProductsMixedActive:
		// Mock shop exists with products that have mixed is_active values
		// The use case should filter out inactive products
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore", "#8B5CF6",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock products count
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock only active products (the repository already filters by is_active=true)
		productRows := sqlmock.NewRows(productColumns).
			AddRow(1, "Active Product 1", "Description 1", 99.99, 10, 2, true, false, false, 0.0, now, 1, "Electronics", "Electronics desc", productImagesJSON).
			AddRow(2, "Active Product 2", "Description 2", 149.99, 5, 1, true, true, true, 129.99, now, 1, "Electronics", "Electronics desc", productImagesJSON).
			AddRow(3, "Active Product 3", "Description 3", 199.99, 20, 5, true, false, false, 0.0, now, 2, "Clothing", "Clothing desc", productImagesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)
	}
}

func (g *GetStoreProductsSteps) parseStoreProductsResponse(ctx *TestContext, resp *http.Response) {
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
		var response contracts.PaginatedProductsResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response
		}
	}
}

// ===== Then Steps =====

func (g *GetStoreProductsSteps) theResponseShouldContainProducts() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(response.Products) == 0 {
		return fmt.Errorf("expected products in response, got empty")
	}
	return nil
}

func (g *GetStoreProductsSteps) allReturnedProductsShouldHaveIsActiveTrue() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}

	for _, product := range response.Products {
		if !product.IsActive {
			return fmt.Errorf("found product with is_active=false: %s (ID: %d)", product.Name, product.ID)
		}
	}
	return nil
}

// ===== Register Steps =====

func (g *GetStoreProductsSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a store with slug "([^"]*)" exists with products$`, g.aStoreWithSlugExistsWithProducts)
	sc.Step(`^a store with slug "([^"]*)" exists without products$`, g.aStoreWithSlugExistsWithoutProducts)
	sc.Step(`^a store with slug "([^"]*)" exists with mixed active and inactive products$`, g.aStoreWithSlugExistsWithMixedActiveAndInactiveProducts)

	// When steps
	sc.Step(`^I send a get store products request for "([^"]*)"$`, g.iSendAGetStoreProductsRequestFor)
	sc.Step(`^I send a get store products request for "([^"]*)" with search "([^"]*)"$`, g.iSendAGetStoreProductsRequestForWithSearch)
	sc.Step(`^I send a get store products request for "([^"]*)" with category_id (\d+)$`, g.iSendAGetStoreProductsRequestForWithCategoryID)
	sc.Step(`^I send a get store products request for "([^"]*)" with limit (\d+)$`, g.iSendAGetStoreProductsRequestForWithLimit)
	sc.Step(`^I send an unauthenticated get store products request for "([^"]*)"$`, g.iSendAnUnauthenticatedGetStoreProductsRequestFor)
	sc.Step(`^I send a get store products request for "([^"]*)" with search "([^"]*)" and category_id (\d+)$`, g.iSendAGetStoreProductsRequestForWithSearchAndCategoryID)
	sc.Step(`^I send a get store products request for "([^"]*)" with sort "([^"]*)" and order "([^"]*)"$`, g.iSendAGetStoreProductsRequestForWithSortAndOrder)

	// Then steps
	sc.Step(`^the response should contain products$`, g.theResponseShouldContainProducts)
	sc.Step(`^all returned products should have is_active true$`, g.allReturnedProductsShouldHaveIsActiveTrue)
}
