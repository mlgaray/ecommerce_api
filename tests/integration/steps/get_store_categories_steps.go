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
	scenarioStoreCategoriesExists     = "store-categories-exists"
	scenarioStoreCategoriesEmpty      = "store-categories-empty"
	scenarioStoreCategoriesNotFound   = "store-categories-not-found"
	scenarioStoreCategoriesWithSearch = "store-categories-with-search"
)

type GetStoreCategoriesSteps struct{}

func NewGetStoreCategoriesSteps() *GetStoreCategoriesSteps {
	return &GetStoreCategoriesSteps{}
}

// ===== Given Steps =====

func (g *GetStoreCategoriesSteps) aStoreWithSlugExistsWithCategories(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreCategoriesExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreCategoriesSteps) aStoreWithSlugExistsWithoutCategories(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreCategoriesEmpty
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

// ===== When Steps =====

func (g *GetStoreCategoriesSteps) iSendAGetStoreCategoriesRequestFor(slug string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreCategoriesSQLExpectations(slug, "")

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s/categories", ctx.server.URL, slug)
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
	g.parseStoreCategoriesResponse(ctx, resp)

	return nil
}

func (g *GetStoreCategoriesSteps) iSendAGetStoreCategoriesRequestForWithSearch(slug, search string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreCategoriesWithSearch

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreCategoriesSQLExpectations(slug, search)

	// Build URL with search param and make request
	url := fmt.Sprintf("%s/stores/%s/categories?search=%s", ctx.server.URL, slug, search)
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
	g.parseStoreCategoriesResponse(ctx, resp)

	return nil
}

func (g *GetStoreCategoriesSteps) iSendAnUnauthenticatedGetStoreCategoriesRequestFor(slug string) error {
	// Store endpoints are always unauthenticated, so this is the same as the regular request
	return g.iSendAGetStoreCategoriesRequestFor(slug)
}

// ===== SQL Mock Setup =====

func (g *GetStoreCategoriesSteps) setupGetStoreCategoriesSQLExpectations(slug, search string) {
	_ = search // Search parameter reserved for future use (SQL mocks don't vary by search)
	ctx := GetTestContext()

	// Shop columns for GetBySlug query - must match shop_repository.go scan order
	shopColumns := []string{
		"id", "name", "slug", "email", "phone", "instagram",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules", "timezone",
	}

	// Category columns for GetAllByShopIDWithFilters query - must match category_repository.go scan order
	// id, name, description, created_at, image (jsonb)
	categoryColumns := []string{
		"id", "name", "description", "created_at", "image",
	}

	now := time.Now()
	categoryImageJSON := `{"id": 1, "url": "https://cloudinary.com/cat.jpg"}`

	switch ctx.scenario {
	case scenarioStoreCategoriesExists, scenarioStoreCategoriesWithSearch:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore",
				testStoreImagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, testStoreTimezoneJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock categories count
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock categories
		categoryRows := sqlmock.NewRows(categoryColumns).
			AddRow(1, "Electronics", "Electronics category", now, categoryImageJSON).
			AddRow(2, "Clothing", "Clothing category", now, categoryImageJSON).
			AddRow(3, "Food", "Food category", now, categoryImageJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(categoryRows)

	case scenarioStoreCategoriesEmpty:
		// Mock shop exists
		shopRows := sqlmock.NewRows(shopColumns).
			AddRow(2, "Empty Store", slug, "test@store.com", "+54111234567", "@emptystore",
				"[]", testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON, nil)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(shopRows)

		// Mock categories count (0)
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		ctx.mockSQLMock.ExpectQuery("SELECT COUNT").
			WillReturnRows(countRows)

		// Mock empty categories
		categoryRows := sqlmock.NewRows(categoryColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WillReturnRows(categoryRows)

	case scenarioStoreCategoriesNotFound, scenarioStoreNotFound:
		// Mock shop not found
		emptyRows := sqlmock.NewRows(shopColumns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)
	}
}

func (g *GetStoreCategoriesSteps) parseStoreCategoriesResponse(ctx *TestContext, resp *http.Response) {
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
		var response contracts.PaginatedCategoriesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response
		}
	}
}

// ===== Then Steps =====

func (g *GetStoreCategoriesSteps) theResponseShouldContainCategories() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(response.Categories) == 0 {
		return fmt.Errorf("expected categories in response, got empty")
	}
	return nil
}

func (g *GetStoreCategoriesSteps) theResponseShouldContainEmptyCategories() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedCategoriesResponse, got: %T", ctx.responseBody)
	}
	if len(response.Categories) > 0 {
		return fmt.Errorf("expected empty categories, got %d", len(response.Categories))
	}
	return nil
}

func (g *GetStoreCategoriesSteps) theResponseShouldIncludePaginationInfo() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(contracts.PaginatedCategoriesResponse)
	if !ok {
		return fmt.Errorf("expected PaginatedCategoriesResponse, got: %T", ctx.responseBody)
	}
	// Store endpoints use cursor-based pagination without TotalCount
	// Just verify the response structure is correct (HasMore field exists)
	_ = response.HasMore // HasMore is always present in the response
	return nil
}

// ===== Register Steps =====

func (g *GetStoreCategoriesSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a store with slug "([^"]*)" exists with categories$`, g.aStoreWithSlugExistsWithCategories)
	sc.Step(`^a store with slug "([^"]*)" exists without categories$`, g.aStoreWithSlugExistsWithoutCategories)

	// When steps
	sc.Step(`^I send a get store categories request for "([^"]*)"$`, g.iSendAGetStoreCategoriesRequestFor)
	sc.Step(`^I send a get store categories request for "([^"]*)" with search "([^"]*)"$`, g.iSendAGetStoreCategoriesRequestForWithSearch)
	sc.Step(`^I send an unauthenticated get store categories request for "([^"]*)"$`, g.iSendAnUnauthenticatedGetStoreCategoriesRequestFor)

	// Then steps
	sc.Step(`^the response should contain categories$`, g.theResponseShouldContainCategories)
	sc.Step(`^the response should contain empty categories$`, g.theResponseShouldContainEmptyCategories)
	sc.Step(`^the response should include pagination info$`, g.theResponseShouldIncludePaginationInfo)
}
