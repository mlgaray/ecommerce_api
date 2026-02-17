package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	scenarioCategoryExists        = "category-exists"
	scenarioCategoryExistsNoImage = "category-exists-no-image"
	scenarioCategoryNotFound      = "category-not-found"
)

type GetCategoryByIDSteps struct{}

func NewGetCategoryByIDSteps() *GetCategoryByIDSteps {
	return &GetCategoryByIDSteps{}
}

// ===== Given Steps =====

func (g *GetCategoryByIDSteps) theUserIsAuthenticatedForCategoryOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GetCategoryByIDSteps) theUserIsNotAuthenticated() error {
	ctx := GetTestContext()
	ctx.authToken = "" // Clear auth token
	return nil
}

func (g *GetCategoryByIDSteps) aCategoryWithIDExists(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryExists
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)
	return nil
}

func (g *GetCategoryByIDSteps) aCategoryWithIDExistsWithoutImage(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryExistsNoImage
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)
	return nil
}

func (g *GetCategoryByIDSteps) aCategoryWithIDDoesNotExist(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioCategoryNotFound
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)
	return nil
}

// ===== When Steps =====

func (g *GetCategoryByIDSteps) iSendAGetCategoryByIDRequestForCategory(categoryID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	g.setupGetCategoryByIDSQLExpectations(categoryID)

	// Build URL and make request
	url := fmt.Sprintf("%s/categories/%d", ctx.server.URL, categoryID)
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
	g.parseResponse(ctx, resp)

	return nil
}

func (g *GetCategoryByIDSteps) iSendAGetCategoryByIDRequestWithInvalidID(invalidID string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Build URL with invalid ID
	url := fmt.Sprintf("%s/categories/%s", ctx.server.URL, invalidID)
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
	g.parseResponse(ctx, resp)

	return nil
}

func (g *GetCategoryByIDSteps) iSendAnUnauthenticatedGetCategoryByIDRequestForCategory(categoryID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Build URL and make request WITHOUT auth header
	url := fmt.Sprintf("%s/categories/%d", ctx.server.URL, categoryID)
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
	g.parseResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (g *GetCategoryByIDSteps) setupGetCategoryByIDSQLExpectations(categoryID int) {
	ctx := GetTestContext()

	// Columns returned by GetByID
	columns := []string{"id", "name", "description", "created_at", "image"}

	now := time.Now()

	switch ctx.scenario {
	case scenarioCategoryExists:
		// Mock category with image
		imageJSON := `{"id": 5, "url": "https://cloudinary.com/image.jpg", "storage_ref": "shop_1/categories/abc123"}`
		rows := sqlmock.NewRows(columns).
			AddRow(categoryID, "Electronics", "Electronic products", now, imageJSON)

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WithArgs(categoryID).
			WillReturnRows(rows)

	case scenarioCategoryExistsNoImage:
		// Mock category without image
		rows := sqlmock.NewRows(columns).
			AddRow(categoryID, "Clothing", "Clothing products", now, "null")

		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WithArgs(categoryID).
			WillReturnRows(rows)

	case scenarioCategoryNotFound:
		// Mock no rows found
		emptyRows := sqlmock.NewRows(columns)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM categories").
			WithArgs(categoryID).
			WillReturnRows(emptyRows)
	}
}

func (g *GetCategoryByIDSteps) parseResponse(ctx *TestContext, resp *http.Response) {
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
		var response responses.GetCategoryResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response.Category
		}
	}
}

// ===== Then Steps =====

func (g *GetCategoryByIDSteps) theResponseShouldContainTheCategoryDetails() error {
	ctx := GetTestContext()
	category, ok := ctx.responseBody.(*responses.CategoryResponse)
	if !ok {
		return fmt.Errorf("expected CategoryResponse, got: %T", ctx.responseBody)
	}
	if category == nil || category.ID == 0 {
		return fmt.Errorf("expected category with ID, got ID 0")
	}
	if category.Name == "" {
		return fmt.Errorf("expected category with name, got empty name")
	}
	return nil
}

func (g *GetCategoryByIDSteps) theResponseShouldContainTheCategoryImage() error {
	ctx := GetTestContext()
	category, ok := ctx.responseBody.(*responses.CategoryResponse)
	if !ok {
		return fmt.Errorf("expected CategoryResponse, got: %T", ctx.responseBody)
	}
	if category.Image == nil {
		return fmt.Errorf("expected category with image, got nil image")
	}
	if category.Image.URL == "" {
		return fmt.Errorf("expected category image with URL, got empty URL")
	}
	return nil
}

func (g *GetCategoryByIDSteps) theResponseShouldNotContainACategoryImage() error {
	ctx := GetTestContext()
	category, ok := ctx.responseBody.(*responses.CategoryResponse)
	if !ok {
		return fmt.Errorf("expected CategoryResponse, got: %T", ctx.responseBody)
	}
	if category.Image != nil {
		return fmt.Errorf("expected category without image, got image with URL: %s", category.Image.URL)
	}
	return nil
}

// ===== Register Steps =====

func (g *GetCategoryByIDSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for category operations$`, g.theUserIsAuthenticatedForCategoryOperations)
	sc.Step(`^the user is not authenticated$`, g.theUserIsNotAuthenticated)
	sc.Step(`^a category with ID (\d+) exists$`, g.aCategoryWithIDExists)
	sc.Step(`^a category with ID (\d+) exists without image$`, g.aCategoryWithIDExistsWithoutImage)
	sc.Step(`^a category with ID (\d+) does not exist$`, g.aCategoryWithIDDoesNotExist)

	// When steps
	sc.Step(`^I send a get category by ID request for category (\d+)$`, g.iSendAGetCategoryByIDRequestForCategory)
	sc.Step(`^I send a get category by ID request with invalid ID "([^"]*)"$`, g.iSendAGetCategoryByIDRequestWithInvalidID)
	sc.Step(`^I send an unauthenticated get category by ID request for category (\d+)$`, g.iSendAnUnauthenticatedGetCategoryByIDRequestForCategory)

	// Then steps
	sc.Step(`^the response should contain the category details$`, g.theResponseShouldContainTheCategoryDetails)
	sc.Step(`^the response should contain the category image$`, g.theResponseShouldContainTheCategoryImage)
	sc.Step(`^the response should not contain a category image$`, g.theResponseShouldNotContainACategoryImage)
}
