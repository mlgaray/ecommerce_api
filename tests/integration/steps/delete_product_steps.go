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
	scenarioDeleteProductSuccess       = "delete-product-success"
	scenarioDeleteProductWithImages    = "delete-product-with-images"
	scenarioDeleteProductNotFound      = "delete-product-not-found"
	scenarioDeleteProductDifferentShop = "delete-product-different-shop"
)

type DeleteProductSteps struct{}

func NewDeleteProductSteps() *DeleteProductSteps {
	return &DeleteProductSteps{}
}

// ===== Given Steps =====

func (d *DeleteProductSteps) theUserIsAuthenticatedForProductDeleteOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DeleteProductSteps) theUserIsNotAuthenticatedForProductDeletion() error {
	ctx := GetTestContext()
	ctx.authToken = "" // Clear auth token
	return nil
}

func (d *DeleteProductSteps) iHaveAProductWithIDThatHasImages(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteProductWithImages

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)

	return nil
}

func (d *DeleteProductSteps) iHaveAProductWithIDThatHasNoImages(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteProductSuccess

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)

	return nil
}

func (d *DeleteProductSteps) iHaveAProductWithIDThatDoesNotExistForDeletion(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteProductNotFound

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)

	return nil
}

func (d *DeleteProductSteps) iHaveAProductWithIDThatBelongsToADifferentShop(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioDeleteProductDifferentShop

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)

	return nil
}

func (d *DeleteProductSteps) iHaveAnInvalidProductIDForDeletion(invalidID string) error {
	ctx := GetTestContext()

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = invalidID

	return nil
}

// ===== When Steps =====

func (d *DeleteProductSteps) iSendADeleteProductRequest() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	d.setupDeleteProductSQLExpectations()

	productID := ctx.pathParams["product_id"]
	return d.sendDeleteProductRequest(productID)
}

func (d *DeleteProductSteps) iSendADeleteProductRequestWithInvalidID() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}

	productID := ctx.pathParams["product_id"]
	return d.sendDeleteProductRequest(productID)
}

func (d *DeleteProductSteps) iSendAnUnauthenticatedDeleteProductRequestForProduct(productID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupProductTestApp(); err != nil {
			return err
		}
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)

	return d.sendDeleteProductRequest(fmt.Sprintf("%d", productID))
}

func (d *DeleteProductSteps) sendDeleteProductRequest(productID string) error {
	ctx := GetTestContext()

	url := fmt.Sprintf("%s/products/%s", ctx.server.URL, productID)
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

func (d *DeleteProductSteps) setupDeleteProductSQLExpectations() {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioDeleteProductSuccess:
		// Mock successful product deletion without images
		// CTE query returns: deleted count = 1, storage_refs = empty array
		rows := sqlmock.NewRows([]string{"deleted", "storage_refs"}).
			AddRow(1, pq.Array([]string{}))
		ctx.mockSQLMock.ExpectQuery("WITH deleted_images AS").
			WillReturnRows(rows)

	case scenarioDeleteProductWithImages:
		// Mock successful product deletion with images
		// CTE query returns: deleted count = 1, storage_refs = array of refs
		rows := sqlmock.NewRows([]string{"deleted", "storage_refs"}).
			AddRow(1, pq.Array([]string{"shop_1/products/image1.jpg", "shop_1/products/image2.jpg"}))
		ctx.mockSQLMock.ExpectQuery("WITH deleted_images AS").
			WillReturnRows(rows)

	case scenarioDeleteProductNotFound:
		// Mock product not found - returns deleted count = 0
		rows := sqlmock.NewRows([]string{"deleted", "storage_refs"}).
			AddRow(0, pq.Array([]string{}))
		ctx.mockSQLMock.ExpectQuery("WITH deleted_images AS").
			WillReturnRows(rows)

	case scenarioDeleteProductDifferentShop:
		// Mock product belongs to different shop - returns deleted count = 0
		// (same as not found, because WHERE shop_id = $2 doesn't match)
		rows := sqlmock.NewRows([]string{"deleted", "storage_refs"}).
			AddRow(0, pq.Array([]string{}))
		ctx.mockSQLMock.ExpectQuery("WITH deleted_images AS").
			WillReturnRows(rows)
	}
}

func (d *DeleteProductSteps) parseResponse(ctx *TestContext, resp *http.Response) {
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

func (d *DeleteProductSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for product delete operations$`, d.theUserIsAuthenticatedForProductDeleteOperations)
	sc.Step(`^the user is not authenticated for product deletion$`, d.theUserIsNotAuthenticatedForProductDeletion)
	sc.Step(`^I have a product with id (\d+) that has images$`, d.iHaveAProductWithIDThatHasImages)
	sc.Step(`^I have a product with id (\d+) that has no images$`, d.iHaveAProductWithIDThatHasNoImages)
	sc.Step(`^I have a product with id (\d+) that does not exist for deletion$`, d.iHaveAProductWithIDThatDoesNotExistForDeletion)
	sc.Step(`^I have a product with id (\d+) that belongs to a different shop$`, d.iHaveAProductWithIDThatBelongsToADifferentShop)
	sc.Step(`^I have an invalid product_id "([^"]*)" for deletion$`, d.iHaveAnInvalidProductIDForDeletion)

	// When steps
	sc.Step(`^I send a delete product request$`, d.iSendADeleteProductRequest)
	sc.Step(`^I send a delete product request with invalid id$`, d.iSendADeleteProductRequestWithInvalidID)
	sc.Step(`^I send an unauthenticated delete product request for product (\d+)$`, d.iSendAnUnauthenticatedDeleteProductRequestForProduct)
}
