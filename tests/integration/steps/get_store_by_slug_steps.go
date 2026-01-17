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
	scenarioStoreExists         = "store-exists"
	scenarioStoreExistsNoImages = "store-exists-no-images"
	scenarioStoreNotFound       = "store-not-found"

	// Test JSON data constants for store
	testStoreAddressJSON         = `{"id": 1, "name": "Main Street 123", "place_id": "ChIJ123", "lat": -34.6037, "lng": -58.3816}`
	testStorePaymentMethodsJSON  = `[{"id": 1, "name": "Transfer", "code": "transfer", "is_active": true}]`
	testStoreDeliveryMethodsJSON = `[{"id": 1, "name": "Delivery", "code": "delivery", "is_active": true}]`
	testStoreImagesJSON          = `[{"id": 1, "url": "https://cloudinary.com/logo.jpg", "type": "logo"}]`
	testStoreImagesFullJSON      = `[{"id": 1, "url": "https://cloudinary.com/logo.jpg", "type": "logo"}, {"id": 2, "url": "https://cloudinary.com/cover.jpg", "type": "cover"}]`
	testProductImagesJSON        = `[{"id": 1, "url": "https://cloudinary.com/product1.jpg", "type": "gallery"}]`
	testEmptySchedulesJSON       = `[]`
	testEmptyVariantsJSON        = `[]`
)

type GetStoreBySlugSteps struct{}

func NewGetStoreBySlugSteps() *GetStoreBySlugSteps {
	return &GetStoreBySlugSteps{}
}

// ===== Given Steps =====

func (g *GetStoreBySlugSteps) aStoreWithSlugExists(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreBySlugSteps) aStoreWithSlugExistsWithoutImages(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreExistsNoImages
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

func (g *GetStoreBySlugSteps) aStoreWithSlugDoesNotExist(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioStoreNotFound
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["slug"] = slug
	return nil
}

// ===== When Steps =====

func (g *GetStoreBySlugSteps) iSendAGetStoreBySlugRequestFor(slug string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetStoreBySlugSQLExpectations(slug)

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/%s", ctx.server.URL, slug)
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
	g.parseStoreResponse(ctx, resp)

	return nil
}

func (g *GetStoreBySlugSteps) iSendAnUnauthenticatedGetStoreBySlugRequestFor(slug string) error {
	// Store endpoints are always unauthenticated, so this is the same as the regular request
	return g.iSendAGetStoreBySlugRequestFor(slug)
}

// ===== SQL Mock Setup =====

func (g *GetStoreBySlugSteps) setupGetStoreBySlugSQLExpectations(slug string) {
	ctx := GetTestContext()

	// Columns returned by GetBySlug query (same as GetByID but uses slug)
	columns := []string{
		"id", "name", "slug", "email", "phone", "instagram", "created_at",
		"images", "address", "payment_methods", "delivery_methods", "operating_schedules",
	}

	now := time.Now()

	switch ctx.scenario {
	case scenarioStoreExists:
		// Mock store with all relations
		imagesJSON := testStoreImagesFullJSON

		rows := sqlmock.NewRows(columns).
			AddRow(1, "Test Store", slug, "test@store.com", "+54111234567", "@teststore", now,
				imagesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testOperatingSchedulesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(rows)

	case scenarioStoreExistsNoImages:
		// Mock store without images
		rows := sqlmock.NewRows(columns).
			AddRow(2, "Test Store No Images", slug, "test@store.com", "+54111234567", "@teststore", now,
				testEmptySchedulesJSON, testStoreAddressJSON, testStorePaymentMethodsJSON, testStoreDeliveryMethodsJSON, testEmptySchedulesJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(rows)

	case scenarioStoreNotFound:
		// Mock no rows found
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(slug).
			WillReturnRows(emptyRows)
	}
}

func (g *GetStoreBySlugSteps) parseStoreResponse(ctx *TestContext, resp *http.Response) {
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
		var store models.Store
		if err := json.NewDecoder(resp.Body).Decode(&store); err == nil {
			ctx.responseBody = store
		}
	}
}

// ===== Then Steps =====

func (g *GetStoreBySlugSteps) theResponseShouldContainTheStoreDetails() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if store.ID == 0 {
		return fmt.Errorf("expected store with ID, got ID 0")
	}
	if store.Name == "" {
		return fmt.Errorf("expected store with name, got empty name")
	}
	if store.Slug == "" {
		return fmt.Errorf("expected store with slug, got empty slug")
	}
	return nil
}

func (g *GetStoreBySlugSteps) theResponseShouldContainTheStoreImages() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if len(store.Images) == 0 {
		return fmt.Errorf("expected store with images, got nil or empty images")
	}
	return nil
}

func (g *GetStoreBySlugSteps) theResponseShouldNotContainStoreImages() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if len(store.Images) > 0 {
		return fmt.Errorf("expected store without images, got %d images", len(store.Images))
	}
	return nil
}

func (g *GetStoreBySlugSteps) theResponseShouldContainTheStoreAddress() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if store.Address == nil {
		return fmt.Errorf("expected store with address, got nil address")
	}
	return nil
}

func (g *GetStoreBySlugSteps) theResponseShouldContainTheStorePaymentMethods() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if len(store.PaymentMethods) == 0 {
		return fmt.Errorf("expected store with payment methods, got nil or empty")
	}
	return nil
}

func (g *GetStoreBySlugSteps) theResponseShouldContainTheStoreDeliveryMethods() error {
	ctx := GetTestContext()
	store, ok := ctx.responseBody.(models.Store)
	if !ok {
		return fmt.Errorf("expected Store, got: %T", ctx.responseBody)
	}
	if len(store.DeliveryMethods) == 0 {
		return fmt.Errorf("expected store with delivery methods, got nil or empty")
	}
	return nil
}

// ===== Register Steps =====

func (g *GetStoreBySlugSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a store with slug "([^"]*)" exists$`, g.aStoreWithSlugExists)
	sc.Step(`^a store with slug "([^"]*)" exists without images$`, g.aStoreWithSlugExistsWithoutImages)
	sc.Step(`^a store with slug "([^"]*)" does not exist$`, g.aStoreWithSlugDoesNotExist)

	// When steps
	sc.Step(`^I send a get store by slug request for "([^"]*)"$`, g.iSendAGetStoreBySlugRequestFor)
	sc.Step(`^I send an unauthenticated get store by slug request for "([^"]*)"$`, g.iSendAnUnauthenticatedGetStoreBySlugRequestFor)

	// Then steps
	sc.Step(`^the response should contain the store details$`, g.theResponseShouldContainTheStoreDetails)
	sc.Step(`^the response should contain the store images$`, g.theResponseShouldContainTheStoreImages)
	sc.Step(`^the response should not contain store images$`, g.theResponseShouldNotContainStoreImages)
	sc.Step(`^the response should contain the store address$`, g.theResponseShouldContainTheStoreAddress)
	sc.Step(`^the response should contain the store payment methods$`, g.theResponseShouldContainTheStorePaymentMethods)
	sc.Step(`^the response should contain the store delivery methods$`, g.theResponseShouldContainTheStoreDeliveryMethods)
}
