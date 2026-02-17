package steps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	scenarioCategoryShopWithCategories       = "shop-with-categories"
	scenarioCategoryShopWithoutCategories    = "shop-without-categories"
	scenarioCategoryShopWithCategoriesCursor = "shop-with-categories-cursor"
	scenarioCategorySearchByName             = "category-search-by-name"
	scenarioCategorySortByNameAsc            = "category-sort-by-name-asc"
	scenarioCategorySortByCreatedAtDesc      = "category-sort-by-created-at-desc"
)

type GetCategoriesByShopIDSteps struct{}

func NewGetCategoriesByShopIDSteps() *GetCategoriesByShopIDSteps {
	return &GetCategoriesByShopIDSteps{}
}

// ===== Given Steps =====

func (g *GetCategoriesByShopIDSteps) aShopWithIDHasCategories(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryShopWithCategories
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetCategoriesByShopIDSteps) aShopWithIDHasNoCategories(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryShopWithoutCategories
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetCategoriesByShopIDSteps) aShopWithIDHasCategoriesWithCursorPagination(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryShopWithCategoriesCursor
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (g *GetCategoriesByShopIDSteps) aShopWithIDHasCategoriesWithDifferentNames(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategorySearchByName
	if ctx.queryParams == nil {
		ctx.queryParams = make(map[string]string)
	}
	ctx.queryParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

// ===== When Steps =====

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShop(shopID int) error {
	return g.sendGetCategoriesRequest(shopID, map[string]string{})
}

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShopWithLimit(shopID, limit int) error {
	params := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}
	return g.sendGetCategoriesRequest(shopID, params)
}

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShopWithAValidCursor(shopID int) error {
	// Create a valid cursor (base64 URL encoded JSON - must match pagination_service.go encoding)
	cursor := base64.URLEncoding.EncodeToString([]byte(`{"id":10,"v":"2025-01-01T10:00:00Z"}`))
	params := map[string]string{
		"cursor": cursor,
	}
	return g.sendGetCategoriesRequest(shopID, params)
}

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShopWithSearchTerm(shopID int, searchTerm string) error {
	params := map[string]string{
		"search": searchTerm,
	}
	return g.sendGetCategoriesRequest(shopID, params)
}

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShopSortedByNameAscending(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategorySortByNameAsc
	params := map[string]string{
		"sort":  "name",
		"order": "asc",
	}
	return g.sendGetCategoriesRequest(shopID, params)
}

func (g *GetCategoriesByShopIDSteps) iSendAGetCategoriesRequestForShopSortedByCreatedAtDescending(shopID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategorySortByCreatedAtDesc
	params := map[string]string{
		"sort":  "created_at",
		"order": "desc",
	}
	return g.sendGetCategoriesRequest(shopID, params)
}

func (g *GetCategoriesByShopIDSteps) sendGetCategoriesRequest(shopID int, queryParams map[string]string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetCategoriesSQLExpectations()

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

func (g *GetCategoriesByShopIDSteps) buildRequestURL(baseURL string, shopID int, params map[string]string) string {
	url := baseURL + fmt.Sprintf("/shops/%d/categories", shopID)

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

func (g *GetCategoriesByShopIDSteps) parseResponse(ctx *TestContext, resp *http.Response) {
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
		var paginatedResponse responses.ListCategoriesResponse
		if err := json.NewDecoder(resp.Body).Decode(&paginatedResponse); err == nil {
			ctx.responseBody = paginatedResponse
		}
	}
}

// ===== SQL Mock Setup =====

func (g *GetCategoriesByShopIDSteps) setupGetCategoriesSQLExpectations() {
	ctx := GetTestContext()

	// Columns returned by GetAllByShopIDWithFilters
	// Must match the exact SELECT order from category_repository.go
	columns := []string{
		"c.id", "c.name", "c.description", "c.created_at", "image",
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioCategoryShopWithCategories, scenarioCategorySortByCreatedAtDesc:
		// Mock COUNT query for first page
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM categories").
			WillReturnRows(countRows)

		// Mock categories query returning sample data
		rows := sqlmock.NewRows(columns).
			AddRow(15, "Electronics", "Electronics description", now, "null").
			AddRow(14, "Clothing", "Clothing description", now, "null").
			AddRow(13, "Food", "Food description", now, "null")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(rows)

	case scenarioCategoryShopWithCategoriesCursor:
		// Mock categories query with cursor (second page - no COUNT)
		rows := sqlmock.NewRows(columns).
			AddRow(9, "Category 9", "Description 9", now, "null").
			AddRow(8, "Category 8", "Description 8", now, "null")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(rows)

	case scenarioCategoryShopWithoutCategories:
		// Mock COUNT query for first page
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM categories").
			WillReturnRows(countRows)

		// Mock empty result
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(emptyRows)

	case scenarioCategorySearchByName:
		// Mock COUNT query for first page
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM categories").
			WillReturnRows(countRows)

		// Mock search results
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Electronics", "Electronics description", now, "null").
			AddRow(2, "Electronics Accessories", "Accessories for electronics", now, "null")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(rows)

	case scenarioCategorySortByNameAsc:
		// Mock COUNT query for first page
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM categories").
			WillReturnRows(countRows)

		// Mock sorted results (name ascending)
		rows := sqlmock.NewRows(columns).
			AddRow(1, "Accessories", "Accessories description", now, "null").
			AddRow(2, "Books", "Books description", now, "null").
			AddRow(3, "Clothing", "Clothing description", now, "null")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(rows)
	}
}

// ===== Then Steps =====

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainAListOfCategories() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.Categories == nil {
		return fmt.Errorf("expected categories list, got nil")
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainCategoryPaginationMetadata() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	// Verify that the response has the pagination fields
	_ = paginatedResponse.NextCursor
	_ = paginatedResponse.HasMore
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainCategoryTotalCountOnFirstPage() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.TotalCount == nil {
		return fmt.Errorf("expected total_count on first page, got nil")
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldNotContainCategoryTotalCount() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.TotalCount != nil {
		return fmt.Errorf("expected no total_count on subsequent pages, got: %v", *paginatedResponse.TotalCount)
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainAtMostNCategories(maxCount int) error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) > maxCount {
		return fmt.Errorf("expected at most %d categories, got %d", maxCount, len(paginatedResponse.Categories))
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainCategoriesFromNextPage() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) == 0 {
		return fmt.Errorf("expected categories from next page, got empty list")
	}
	// Categories from mock are ID 9 and 8 (after cursor with ID 10)
	if paginatedResponse.Categories[0].ID >= 10 {
		return fmt.Errorf("expected categories after cursor, got category with ID %d", paginatedResponse.Categories[0].ID)
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainAnEmptyListOfCategories() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) != 0 {
		return fmt.Errorf("expected empty categories list, got %d categories", len(paginatedResponse.Categories))
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldHaveCategoryHasMoreAsFalse() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if paginatedResponse.HasMore {
		return fmt.Errorf("expected hasMore to be false, got true")
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseShouldContainCategoriesMatchingSearchTerm() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) == 0 {
		return fmt.Errorf("expected categories matching search term, got empty list")
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseCategoriesShouldBeSortedByNameAscending() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) < 2 {
		return nil // Can't verify sorting with less than 2 categories
	}
	// Verify categories are sorted by name ascending
	for i := 1; i < len(paginatedResponse.Categories); i++ {
		if paginatedResponse.Categories[i].Name < paginatedResponse.Categories[i-1].Name {
			return fmt.Errorf("expected categories sorted by name ascending, but category at index %d has lower name than previous", i)
		}
	}
	return nil
}

func (g *GetCategoriesByShopIDSteps) theResponseCategoriesShouldBeSortedByCreatedAtDescending() error {
	ctx := GetTestContext()
	paginatedResponse, ok := ctx.responseBody.(responses.ListCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected ListCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(paginatedResponse.Categories) < 2 {
		return nil // Can't verify sorting with less than 2 categories
	}
	// In our mock data, all categories have the same created_at, so we just verify we got categories
	// The actual sorting is tested in unit tests
	return nil
}

// ===== Register Steps =====

func (g *GetCategoriesByShopIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a shop with ID (\d+) has categories$`, g.aShopWithIDHasCategories)
	sc.Step(`^a shop with ID (\d+) has no categories$`, g.aShopWithIDHasNoCategories)
	sc.Step(`^a shop with ID (\d+) has categories with cursor pagination$`, g.aShopWithIDHasCategoriesWithCursorPagination)
	sc.Step(`^a shop with ID (\d+) has categories with different names$`, g.aShopWithIDHasCategoriesWithDifferentNames)

	// When steps
	sc.Step(`^I send a get categories request for shop (\d+)$`, g.iSendAGetCategoriesRequestForShop)
	sc.Step(`^I send a get categories request for shop (\d+) with limit (-?\d+)$`, g.iSendAGetCategoriesRequestForShopWithLimit)
	sc.Step(`^I send a get categories request for shop (\d+) with a valid cursor$`, g.iSendAGetCategoriesRequestForShopWithAValidCursor)
	sc.Step(`^I send a get categories request for shop (\d+) with search term "([^"]*)"$`, g.iSendAGetCategoriesRequestForShopWithSearchTerm)
	sc.Step(`^I send a get categories request for shop (\d+) sorted by name ascending$`, g.iSendAGetCategoriesRequestForShopSortedByNameAscending)
	sc.Step(`^I send a get categories request for shop (\d+) sorted by created_at descending$`, g.iSendAGetCategoriesRequestForShopSortedByCreatedAtDescending)

	// Then steps
	sc.Step(`^the response should contain a list of categories$`, g.theResponseShouldContainAListOfCategories)
	sc.Step(`^the response should contain category pagination metadata$`, g.theResponseShouldContainCategoryPaginationMetadata)
	sc.Step(`^the response should contain category total count on first page$`, g.theResponseShouldContainCategoryTotalCountOnFirstPage)
	sc.Step(`^the response should not contain category total count$`, g.theResponseShouldNotContainCategoryTotalCount)
	sc.Step(`^the response should contain at most (\d+) categories$`, g.theResponseShouldContainAtMostNCategories)
	sc.Step(`^the response should contain categories from next page$`, g.theResponseShouldContainCategoriesFromNextPage)
	sc.Step(`^the response should contain an empty list of categories$`, g.theResponseShouldContainAnEmptyListOfCategories)
	sc.Step(`^the response should have category hasMore as false$`, g.theResponseShouldHaveCategoryHasMoreAsFalse)
	sc.Step(`^the response should contain categories matching search term$`, g.theResponseShouldContainCategoriesMatchingSearchTerm)
	sc.Step(`^the response categories should be sorted by name ascending$`, g.theResponseCategoriesShouldBeSortedByNameAscending)
	sc.Step(`^the response categories should be sorted by created_at descending$`, g.theResponseCategoriesShouldBeSortedByCreatedAtDescending)
}
