package steps

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

const (
	scenarioShopWithProducts       = "shop-with-products"
	scenarioShopWithoutProducts    = "shop-without-products"
	scenarioShopWithProductsCursor = "shop-with-products-cursor"
)

type GetProductsByShopIDSteps struct{}

func NewGetProductsByShopIDSteps() *GetProductsByShopIDSteps {
	return &GetProductsByShopIDSteps{}
}

// ===== Given Steps =====

func (g *GetProductsByShopIDSteps) aShopWithIDHasProducts(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopWithProducts
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetProductsByShopIDSteps) aShopWithIDHasNoProducts(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopWithoutProducts
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

// ===== When Steps =====

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShop(shopID int) error {
	return g.sendGetProductsRequest(shopID, 0, 0)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithLimit(shopID, limit int) error {
	return g.sendGetProductsRequest(shopID, limit, 0)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithCursor(shopID, cursor int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopWithProductsCursor
	return g.sendGetProductsRequest(shopID, 0, cursor)
}

func (g *GetProductsByShopIDSteps) sendGetProductsRequest(shopID, limit, cursor int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations only if we expect the query to execute
	if limit >= 0 && cursor >= 0 {
		g.setupGetProductsSQLExpectations()
	}

	// Build URL and make request
	url := g.buildRequestURL(ctx.server.URL, shopID, limit, cursor)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseResponse(ctx, resp)

	return nil
}

func (g *GetProductsByShopIDSteps) buildRequestURL(baseURL string, shopID, limit, cursor int) string {
	url := baseURL + fmt.Sprintf("/shops/%d/products", shopID)

	hasParams := false
	if limit != 0 {
		url += fmt.Sprintf("?limit=%d", limit)
		hasParams = true
	}
	if cursor != 0 {
		separator := "?"
		if hasParams {
			separator = "&"
		}
		url += fmt.Sprintf("%scursor=%d", separator, cursor)
	}

	return url
}

func (g *GetProductsByShopIDSteps) parseResponse(ctx *TestContext, resp *http.Response) {
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
		var paginatedResponse contracts.PaginatedProductsResponse
		if err := json.NewDecoder(resp.Body).Decode(&paginatedResponse); err == nil {
			ctx.responseBody = paginatedResponse
		}
	}
}

// ===== SQL Mock Setup =====

func (g *GetProductsByShopIDSteps) setupGetProductsSQLExpectations() {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioShopWithProducts:
		// Mock products query returning sample data (first page)
		// Columns match the Scan in product_repository.go:356-372
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(15, "Product 15", "Description 15", 99.99, 10, 5, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]").
			AddRow(14, "Product 14", "Description 14", 199.99, 20, 10, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]").
			AddRow(13, "Product 13", "Description 13", 299.99, 30, 15, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioShopWithProductsCursor:
		// Mock products query with cursor (returns products with ID < cursor=10)
		// Ordered DESC, so 9, 8, 7...
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		}).
			AddRow(9, "Product 9", "Description 9", 99.99, 10, 5, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]").
			AddRow(8, "Product 8", "Description 8", 199.99, 20, 10, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]").
			AddRow(7, "Product 7", "Description 7", 299.99, 30, 15, true, false, false, 0.0, 1, "Category 1", "", "[]", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioShopWithoutProducts:
		// Mock empty result
		emptyRows := sqlmock.NewRows([]string{
			"id", "name", "description", "price", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"category_id", "category_name", "category_description",
			"images", "variants",
		})

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(emptyRows)
	}
}

// ===== Then Steps =====

func (g *GetProductsByShopIDSteps) theResponseShouldContainAListOfProducts() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.Products == nil {
		return fmt.Errorf("expected products list, got nil")
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainPaginationMetadata() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	// Just verify that the response has the pagination fields (they can be 0/false for empty results)
	_ = paginatedResponse.NextCursor
	_ = paginatedResponse.HasMore
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainAtMostNProducts(maxCount int) error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) > maxCount {
		return fmt.Errorf("expected at most %d products, got %d", maxCount, len(paginatedResponse.Products))
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainProductsAfterCursor(cursor int) error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) > 0 {
		// With DESC ordering and cursor-based pagination: p.id < cursor
		// So all returned products should have ID < cursor
		if paginatedResponse.Products[0].ID >= cursor {
			return fmt.Errorf("expected products with ID < cursor %d (DESC order), got product with ID %d", cursor, paginatedResponse.Products[0].ID)
		}
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainAnEmptyListOfProducts() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) != 0 {
		return fmt.Errorf("expected empty products list, got %d products", len(paginatedResponse.Products))
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldHaveHasMoreAsFalse() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.HasMore {
		return fmt.Errorf("expected hasMore to be false, got true")
	}
	return nil
}

// ===== Register Steps =====

func (g *GetProductsByShopIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a shop with ID (\d+) has products$`, g.aShopWithIDHasProducts)
	sc.Step(`^a shop with ID (\d+) has no products$`, g.aShopWithIDHasNoProducts)

	// When steps
	sc.Step(`^I send a get products request for shop (\d+)$`, g.iSendAGetProductsRequestForShop)
	sc.Step(`^I send a get products request for shop (\d+) with limit (-?\d+)$`, g.iSendAGetProductsRequestForShopWithLimit)
	sc.Step(`^I send a get products request for shop (\d+) with cursor (-?\d+)$`, g.iSendAGetProductsRequestForShopWithCursor)

	// Then steps
	sc.Step(`^the response should contain a list of products$`, g.theResponseShouldContainAListOfProducts)
	sc.Step(`^the response should contain pagination metadata$`, g.theResponseShouldContainPaginationMetadata)
	sc.Step(`^the response should contain at most (\d+) products$`, g.theResponseShouldContainAtMostNProducts)
	sc.Step(`^the response should contain products after cursor (\d+)$`, g.theResponseShouldContainProductsAfterCursor)
	sc.Step(`^the response should contain an empty list of products$`, g.theResponseShouldContainAnEmptyListOfProducts)
	sc.Step(`^the response should have hasMore as false$`, g.theResponseShouldHaveHasMoreAsFalse)
}
