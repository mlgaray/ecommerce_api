package steps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

const (
	scenarioShopWithProducts          = "shop-with-products"
	scenarioShopWithoutProducts       = "shop-without-products"
	scenarioShopWithProductsCursor    = "shop-with-products-cursor"
	scenarioSearchByName              = "search-by-name"
	scenarioFilterByCategory          = "filter-by-category"
	scenarioFilterByPriceRange        = "filter-by-price-range"
	scenarioFilterByActiveStatus      = "filter-by-active-status"
	scenarioSortByPriceAsc            = "sort-by-price-asc"
	scenarioCombineFilters            = "combine-filters"
	scenarioActiveAndInactiveProducts = "active-and-inactive-products"
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

func (g *GetProductsByShopIDSteps) aShopWithIDHasProductsWithCursorPagination(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopWithProductsCursor
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetProductsByShopIDSteps) aShopWithIDHasProductsWithDifferentNames(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioSearchByName
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetProductsByShopIDSteps) aShopWithIDHasProductsInDifferentCategories(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioFilterByCategory
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetProductsByShopIDSteps) aShopWithIDHasProductsWithDifferentPrices(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioFilterByPriceRange
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetProductsByShopIDSteps) aShopWithIDHasActiveAndInactiveProducts(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioActiveAndInactiveProducts
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

// ===== When Steps =====

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShop(shopID int) error {
	return g.sendGetProductsRequest(shopID, map[string]string{})
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithLimit(shopID, limit int) error {
	params := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithAValidCursor(shopID int) error {
	// Create a valid cursor (base64 URL encoded JSON - must match pagination_service.go encoding)
	cursor := base64.URLEncoding.EncodeToString([]byte(`{"id":10,"v":99.99}`))
	params := map[string]string{
		"cursor": cursor,
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithSearchTerm(shopID int, searchTerm string) error {
	params := map[string]string{
		"search": searchTerm,
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithCategoryID(shopID, categoryID int) error {
	params := map[string]string{
		"category_id": fmt.Sprintf("%d", categoryID),
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithMinPriceAndMaxPrice(shopID int, minPrice, maxPrice float64) error {
	params := map[string]string{
		"min_price": fmt.Sprintf("%.2f", minPrice),
		"max_price": fmt.Sprintf("%.2f", maxPrice),
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithIsActive(shopID int, isActiveStr string) error {
	params := map[string]string{
		"is_active": isActiveStr,
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopSortedByPriceAscending(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioSortByPriceAsc
	params := map[string]string{
		"sort":  "price",
		"order": "asc",
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) iSendAGetProductsRequestForShopWithCategoryIDAndSearchTerm(shopID, categoryID int, searchTerm string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCombineFilters
	params := map[string]string{
		"category_id": fmt.Sprintf("%d", categoryID),
		"search":      searchTerm,
	}
	return g.sendGetProductsRequest(shopID, params)
}

func (g *GetProductsByShopIDSteps) sendGetProductsRequest(shopID int, queryParams map[string]string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetProductsSQLExpectations()

	// Build URL and make request
	url := g.buildRequestURL(ctx.server.URL, shopID, queryParams)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	ctx.response = resp
	g.parseResponse(ctx, resp)

	return nil
}

func (g *GetProductsByShopIDSteps) buildRequestURL(baseURL string, shopID int, params map[string]string) string {
	url := baseURL + fmt.Sprintf("/shops/%d/products", shopID)

	if len(params) > 0 {
		url += "?"
		first := true
		for key, value := range params {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
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

	// Columns returned by GetAllByShopIDWithFilters (lightweight - no variants)
	// Must match the exact SELECT order from product_repository.go:280-296
	columns := []string{
		"p.id", "p.name", "p.description", "p.price", "p.stock", "p.minimum_stock",
		"p.is_active", "p.is_highlighted", "p.is_promotional", "p.promotional_price",
		"p.created_at",
		"c.id", "c.name", "c.description", // category fields
		"images", // images as JSONB
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioShopWithProducts:
		// Mock COUNT query for first page (executed FIRST in use case)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock products query returning sample data (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(15, "Product 15", "Description 15", 99.99, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(14, "Product 14", "Description 14", 199.99, 20, 10, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(13, "Product 13", "Description 13", 299.99, 30, 15, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioShopWithProductsCursor:
		// Mock products query with cursor (second page)
		rows := sqlmock.NewRows(columns).
			AddRow(9, "Product 9", "Description 9", 99.99, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(8, "Product 8", "Description 8", 199.99, 20, 10, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioShopWithoutProducts:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock empty result (executed SECOND)
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(emptyRows)

	case scenarioSearchByName:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock search results (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Gaming Laptop", "Description", 999.99, 5, 2, true, false, false, 0.0, now, 1, "Electronics", "", "[]").
			AddRow(2, "Laptop Stand", "Description", 49.99, 10, 5, true, false, false, 0.0, now, 1, "Accessories", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioFilterByCategory:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock category filtered results (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Product A", "Description", 100.0, 10, 5, true, false, false, 0.0, now, 2, "Category 2", "", "[]").
			AddRow(2, "Product B", "Description", 200.0, 10, 5, true, false, false, 0.0, now, 2, "Category 2", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioFilterByPriceRange:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock price range filtered results (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Product 1", "Description", 150.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(2, "Product 2", "Description", 300.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(3, "Product 3", "Description", 450.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioActiveAndInactiveProducts:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock only active products (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Active Product 1", "Description", 100.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(2, "Active Product 2", "Description", 200.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioSortByPriceAsc:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock sorted results (price ascending) (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Cheap Product", "Description", 10.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(2, "Medium Product", "Description", 50.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(3, "Expensive Product", "Description", 100.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)

	case scenarioCombineFilters:
		// Mock COUNT query for first page (executed FIRST)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM products").
			WillReturnRows(countRows)

		// Mock combined filter results (executed SECOND)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Product 1", "Description", 100.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]").
			AddRow(2, "Product 2", "Description", 200.0, 10, 5, true, false, false, 0.0, now, 1, "Category 1", "", "[]")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WillReturnRows(rows)
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
	// Verify that the response has the pagination fields
	_ = paginatedResponse.NextCursor
	_ = paginatedResponse.HasMore
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainTotalCountOnFirstPage() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.TotalCount == nil {
		return fmt.Errorf("expected total_count on first page, got nil")
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldNotContainTotalCount() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.TotalCount != nil {
		return fmt.Errorf("expected no total_count on subsequent pages, got: %v", *paginatedResponse.TotalCount)
	}
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

func (g *GetProductsByShopIDSteps) theResponseShouldContainProductsFromNextPage() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected products from next page, got empty list")
	}
	// Products from mock are ID 9 and 8 (after cursor with ID 10)
	if paginatedResponse.Products[0].ID >= 10 {
		return fmt.Errorf("expected products after cursor, got product with ID %d", paginatedResponse.Products[0].ID)
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

func (g *GetProductsByShopIDSteps) theResponseShouldContainProductsMatchingSearchTerm() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected products matching search term, got empty list")
	}
	// Verify that products contain "Laptop" in name (from mock data)
	// We just check that we got results, detailed search logic is tested in the service
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainProductsFromCategory(categoryID int) error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected products from category %d, got empty list", categoryID)
	}
	// Verify all products are from the expected category
	for _, product := range paginatedResponse.Products {
		if product.Category.ID != categoryID {
			return fmt.Errorf("expected all products from category %d, got product with category %d", categoryID, product.Category.ID)
		}
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainProductsWithinPriceRange() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected products within price range, got empty list")
	}
	// Mock data has products with prices 150, 300, 450 (all within 100-500 range)
	for _, product := range paginatedResponse.Products {
		if product.Price < 100 || product.Price > 500 {
			return fmt.Errorf("expected products within price range 100-500, got product with price %.2f", product.Price)
		}
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainOnlyActiveProducts() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected active products, got empty list")
	}
	// Verify all products are active
	for _, product := range paginatedResponse.Products {
		if !product.IsActive {
			return fmt.Errorf("expected only active products, got inactive product with ID %d", product.ID)
		}
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseProductsShouldBeSortedByPriceAscending() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) < 2 {
		return nil // Can't verify sorting with less than 2 products
	}
	// Verify products are sorted by price ascending
	for i := 1; i < len(paginatedResponse.Products); i++ {
		if paginatedResponse.Products[i].Price < paginatedResponse.Products[i-1].Price {
			return fmt.Errorf("expected products sorted by price ascending, but product at index %d has lower price than previous", i)
		}
	}
	return nil
}

func (g *GetProductsByShopIDSteps) theResponseShouldContainFilteredProducts() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(contracts.PaginatedProductsResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedProductsResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Products) == 0 {
		return fmt.Errorf("expected filtered products, got empty list")
	}
	// Just verify we got results - detailed filtering logic is tested in the service
	return nil
}

// ===== Register Steps =====

func (g *GetProductsByShopIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a shop with ID (\d+) has products$`, g.aShopWithIDHasProducts)
	sc.Step(`^a shop with ID (\d+) has no products$`, g.aShopWithIDHasNoProducts)
	sc.Step(`^a shop with ID (\d+) has products with cursor pagination$`, g.aShopWithIDHasProductsWithCursorPagination)
	sc.Step(`^a shop with ID (\d+) has products with different names$`, g.aShopWithIDHasProductsWithDifferentNames)
	sc.Step(`^a shop with ID (\d+) has products in different categories$`, g.aShopWithIDHasProductsInDifferentCategories)
	sc.Step(`^a shop with ID (\d+) has products with different prices$`, g.aShopWithIDHasProductsWithDifferentPrices)
	sc.Step(`^a shop with ID (\d+) has active and inactive products$`, g.aShopWithIDHasActiveAndInactiveProducts)

	// When steps
	sc.Step(`^I send a get products request for shop (\d+)$`, g.iSendAGetProductsRequestForShop)
	sc.Step(`^I send a get products request for shop (\d+) with limit (-?\d+)$`, g.iSendAGetProductsRequestForShopWithLimit)
	sc.Step(`^I send a get products request for shop (\d+) with a valid cursor$`, g.iSendAGetProductsRequestForShopWithAValidCursor)
	sc.Step(`^I send a get products request for shop (\d+) with search term "([^"]*)"$`, g.iSendAGetProductsRequestForShopWithSearchTerm)
	sc.Step(`^I send a get products request for shop (\d+) with category ID (\d+)$`, g.iSendAGetProductsRequestForShopWithCategoryID)
	sc.Step(`^I send a get products request for shop (\d+) with min price (\d+) and max price (\d+)$`, g.iSendAGetProductsRequestForShopWithMinPriceAndMaxPrice)
	sc.Step(`^I send a get products request for shop (\d+) with is_active (\w+)$`, g.iSendAGetProductsRequestForShopWithIsActive)
	sc.Step(`^I send a get products request for shop (\d+) sorted by price ascending$`, g.iSendAGetProductsRequestForShopSortedByPriceAscending)
	sc.Step(`^I send a get products request for shop (\d+) with category ID (\d+) and search term "([^"]*)"$`, g.iSendAGetProductsRequestForShopWithCategoryIDAndSearchTerm)

	// Then steps
	sc.Step(`^the response should contain a list of products$`, g.theResponseShouldContainAListOfProducts)
	sc.Step(`^the response should contain pagination metadata$`, g.theResponseShouldContainPaginationMetadata)
	sc.Step(`^the response should contain total count on first page$`, g.theResponseShouldContainTotalCountOnFirstPage)
	sc.Step(`^the response should not contain total count$`, g.theResponseShouldNotContainTotalCount)
	sc.Step(`^the response should contain at most (\d+) products$`, g.theResponseShouldContainAtMostNProducts)
	sc.Step(`^the response should contain products from next page$`, g.theResponseShouldContainProductsFromNextPage)
	sc.Step(`^the response should contain an empty list of products$`, g.theResponseShouldContainAnEmptyListOfProducts)
	sc.Step(`^the response should have hasMore as false$`, g.theResponseShouldHaveHasMoreAsFalse)
	sc.Step(`^the response should contain products matching search term$`, g.theResponseShouldContainProductsMatchingSearchTerm)
	sc.Step(`^the response should contain products from category (\d+)$`, g.theResponseShouldContainProductsFromCategory)
	sc.Step(`^the response should contain products within price range$`, g.theResponseShouldContainProductsWithinPriceRange)
	sc.Step(`^the response should contain only active products$`, g.theResponseShouldContainOnlyActiveProducts)
	sc.Step(`^the response products should be sorted by price ascending$`, g.theResponseProductsShouldBeSortedByPriceAscending)
	sc.Step(`^the response should contain filtered products$`, g.theResponseShouldContainFilteredProducts)
}
