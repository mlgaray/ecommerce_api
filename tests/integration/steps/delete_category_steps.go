package steps

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
	"github.com/lib/pq"
)

const (
	scenarioDeleteCategorySuccess    = "delete-category-success"
	scenarioDeleteCategoryNotFound   = "delete-category-not-found"
	scenarioDeleteCategoryHasProduct = "delete-category-has-products"
)

type DeleteCategorySteps struct{}

func NewDeleteCategorySteps() *DeleteCategorySteps {
	return &DeleteCategorySteps{}
}

// ===== Given Steps =====

func (d *DeleteCategorySteps) theUserIsAuthenticatedForCategoryDeleteOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DeleteCategorySteps) theUserIsNotAuthenticatedForCategoryDeletion() error {
	ctx := GetTestContext()
	ctx.authToken = "" // Clear auth token
	return nil
}

func (d *DeleteCategorySteps) iHaveACategoryWithIDThatHasNoProducts(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteCategorySuccess

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (d *DeleteCategorySteps) iHaveACategoryWithIDThatHasProductsAssigned(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteCategoryHasProduct

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (d *DeleteCategorySteps) iHaveACategoryWithIDThatDoesNotExistForDeletion(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteCategoryNotFound

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (d *DeleteCategorySteps) iHaveAnInvalidCategoryIDForDeletion(invalidID string) error {
	ctx := GetTestContext()

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = invalidID

	return nil
}

// ===== When Steps =====

func (d *DeleteCategorySteps) iSendADeleteCategoryRequest() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	d.setupDeleteCategorySQLExpectations()

	categoryID := ctx.pathParams["category_id"]
	return d.sendDeleteCategoryRequest(categoryID)
}

func (d *DeleteCategorySteps) iSendADeleteCategoryRequestWithInvalidID() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	categoryID := ctx.pathParams["category_id"]
	return d.sendDeleteCategoryRequest(categoryID)
}

func (d *DeleteCategorySteps) iSendAnUnauthenticatedDeleteCategoryRequestForCategory(categoryID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return d.sendDeleteCategoryRequest(fmt.Sprintf("%d", categoryID))
}

func (d *DeleteCategorySteps) sendDeleteCategoryRequest(categoryID string) error {
	ctx := GetTestContext()

	url := fmt.Sprintf("%s/categories/%s", ctx.server.URL, categoryID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	d.parseResponse(ctx, resp)

	return nil
}

func (d *DeleteCategorySteps) setupDeleteCategorySQLExpectations() {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioDeleteCategorySuccess:
		// Mock successful category deletion
		// CTE query returns: deleted count = 1, storage_ref = "" (no image)
		rows := sqlmock.NewRows([]string{"deleted", "storage_ref"}).
			AddRow(1, "")
		ctx.mockSQLMock.ExpectQuery("WITH deleted_image AS").
			WillReturnRows(rows)

	case scenarioDeleteCategoryNotFound:
		// Mock category not found - returns deleted count = 0
		rows := sqlmock.NewRows([]string{"deleted", "storage_ref"}).
			AddRow(0, "")
		ctx.mockSQLMock.ExpectQuery("WITH deleted_image AS").
			WillReturnRows(rows)

	case scenarioDeleteCategoryHasProduct:
		// Mock FK constraint violation error
		pqErr := &pq.Error{
			Code:    "23503", // foreign_key_violation
			Message: "update or delete on table \"categories\" violates foreign key constraint",
		}
		ctx.mockSQLMock.ExpectQuery("WITH deleted_image AS").
			WillReturnError(pqErr)
	}
}

func (d *DeleteCategorySteps) parseResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	// For 204 No Content, there's no body to parse
	if resp.StatusCode == http.StatusNoContent {
		return
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
		if msg, ok := response["message"]; ok {
			ctx.successMessage = msg
		}
		if errMsg, ok := response["error"]; ok {
			ctx.errorMessage = errMsg
		}
	}
}

// ===== Register Steps =====

func (d *DeleteCategorySteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for category delete operations$`, d.theUserIsAuthenticatedForCategoryDeleteOperations)
	sc.Step(`^the user is not authenticated for category deletion$`, d.theUserIsNotAuthenticatedForCategoryDeletion)
	sc.Step(`^I have a category with id (\d+) that has no products$`, d.iHaveACategoryWithIDThatHasNoProducts)
	sc.Step(`^I have a category with id (\d+) that has products assigned$`, d.iHaveACategoryWithIDThatHasProductsAssigned)
	sc.Step(`^I have a category with id (\d+) that does not exist for deletion$`, d.iHaveACategoryWithIDThatDoesNotExistForDeletion)
	sc.Step(`^I have an invalid category_id "([^"]*)" for deletion$`, d.iHaveAnInvalidCategoryIDForDeletion)

	// When steps
	sc.Step(`^I send a delete category request$`, d.iSendADeleteCategoryRequest)
	sc.Step(`^I send a delete category request with invalid id$`, d.iSendADeleteCategoryRequestWithInvalidID)
	sc.Step(`^I send an unauthenticated delete category request for category (\d+)$`, d.iSendAnUnauthenticatedDeleteCategoryRequestForCategory)
}
