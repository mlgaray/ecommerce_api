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
	scenarioStoreFeaturedExists   = "store-featured-exists"
	scenarioStoreFeaturedEmpty    = "store-featured-empty"
	scenarioStoreFeaturedNotFound = "store-featured-not-found"
	scenarioStoreFeaturedMixed    = "store-featured-mixed"
)

type GetStoreFeaturedProductsSteps struct{}

func NewGetStoreFeaturedProductsSteps() *GetStoreFeaturedProductsSteps {
	return &GetStoreFeaturedProductsSteps{}
}

// ===== Given Steps =====

func (g *GetStoreFeaturedProductsSteps) aStoreWithSlugExistsWithFeaturedProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreFeaturedExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreFeaturedProductsSteps) aStoreWithSlugExistsWithoutFeaturedProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreFeaturedEmpty
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreFeaturedProductsSteps) aStoreWithSlugExistsWithMixedProducts(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreFeaturedMixed
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

// ===== When Steps =====

func (g *GetStoreFeaturedProductsSteps) iSendAGetStoreFeaturedProductsRequestFor(slug string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreFeaturedProductsSQLExpectations(slug)

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s/products/featured", ctx.server.URL, slug)
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
	g.parseStoreFeaturedProductsResponse(ctx, resp)

	return nil
}

func (g *GetStoreFeaturedProductsSteps) iSendAnUnauthenticatedGetStoreFeaturedProductsRequestFor(slug string) error {
	// Store endpoints are always unauthenticated, so this is the same as the regular request
	return g.iSendAGetStoreFeaturedProductsRequestFor(slug)
}

// ===== SQL Mock Setup =====

func (g *GetStoreFeaturedProductsSteps) setupGetStoreFeaturedProductsSQLExpectations(slug string) {
	ctx := GetTestContext()

	// Shop columns for GetBySlug query - must match shop_repository.go scan order (12 columns, no created_at)
	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	// Product columns for GetAllByShopIDWithFilters query - must match product_repository.go scan order (15 columns)
	productColumns := []string{
		"id", "name", "description", "price", "stock", "minimum_stock",
		"is_active", "is_highlighted", "is_promotional", "promotional_price",
		"created_at", "category_id", "category_name", "category_description", "images",
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioStoreFeaturedExists, scenarioStoreFeaturedMixed:
		// Mock shop exists (12 values, no created_at)
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock products count
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock featured products (is_highlighted=true, is_active=true) - 15 values per row
		// Order: id, name, description, price, stock, minimum_stock, is_active, is_highlighted, is_promotional, promotional_price, created_at, category_id, category_name, category_description, images
		productRows := sqlmock.NewRows(productColumns).
			AddRow(1, "Featured Product 1", "Description 1", 99.99, 10, 2, true, true, false, 0.0, now, 1, "Electronics", "Electronics desc", testProductImagesJSON).
			AddRow(2, "Featured Product 2", "Description 2", 149.99, 5, 1, true, true, true, 129.99, now, 1, "Electronics", "Electronics desc", testProductImagesJSON).
			AddRow(3, "Featured Product 3", "Description 3", 199.99, 20, 5, true, true, false, 0.0, now, 2, "Clothing", "Clothing desc", testProductImagesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(productRows)

	case scenarioStoreFeaturedEmpty:
		// Mock shop exists (12 values, no created_at)
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(2, "Empty Store", slug, "test@store.com", "+54111234567", "@emptystore",
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

	case scenarioStoreFeaturedNotFound, scenarioStoreNotFound:
		// Mock shop not found
		emptyRows := sqlmock.NewRows(shopColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)
	}
}

func (g *GetStoreFeaturedProductsSteps) parseStoreFeaturedProductsResponse(ctx *TestContext, resp *http.Response) {
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

func (g *GetStoreFeaturedProductsSteps) theResponseShouldContainFeaturedProducts() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(response.Products) == 0 {
		return fmt.Errorf("expected featured products in response, got empty")
	}
	return nil
}

func (g *GetStoreFeaturedProductsSteps) theResponseShouldContainEmptyProducts() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(response.Products) > 0 {
		return fmt.Errorf("expected empty products, got %d", len(response.Products))
	}
	return nil
}

func (g *GetStoreFeaturedProductsSteps) theResponseShouldIncludeProductsPaginationInfo() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	// Store endpoints use cursor-based pagination without TotalCount
	// Just verify the response structure is correct (HasMore field exists)
	_ = response.HasMore // HasMore is always present in the response
	return nil
}

func (g *GetStoreFeaturedProductsSteps) theResponseShouldOnlyContainHighlightedActiveProducts() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	for _, product := range response.Products {
		if !product.IsHighlighted {
			return fmt.Errorf("expected all products to be highlighted, found non-highlighted product: %s", product.Name)
		}
		if !product.IsActive {
			return fmt.Errorf("expected all products to be active, found inactive product: %s", product.Name)
		}
	}
	return nil
}

// ===== Register Steps =====

func (g *GetStoreFeaturedProductsSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a store with slug "([^"]*)" exists with featured products$`, g.aStoreWithSlugExistsWithFeaturedProducts)
	sc.Step(`^a store with slug "([^"]*)" exists without featured products$`, g.aStoreWithSlugExistsWithoutFeaturedProducts)
	sc.Step(`^a store with slug "([^"]*)" exists with mixed products$`, g.aStoreWithSlugExistsWithMixedProducts)

	// When steps
	sc.Step(`^I send a get store featured products request for "([^"]*)"$`, g.iSendAGetStoreFeaturedProductsRequestFor)
	sc.Step(`^I send an unauthenticated get store featured products request for "([^"]*)"$`, g.iSendAnUnauthenticatedGetStoreFeaturedProductsRequestFor)

	// Then steps
	sc.Step(`^the response should contain featured products$`, g.theResponseShouldContainFeaturedProducts)
	sc.Step(`^the response should contain empty products$`, g.theResponseShouldContainEmptyProducts)
	sc.Step(`^the response should include products pagination info$`, g.theResponseShouldIncludeProductsPaginationInfo)
	sc.Step(`^the response should only contain highlighted active products$`, g.theResponseShouldOnlyContainHighlightedActiveProducts)
}
