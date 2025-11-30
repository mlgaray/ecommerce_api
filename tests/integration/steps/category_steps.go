package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

const (
	validCategoryScenario     = "valid-category"
	duplicateCategoryScenario = "duplicate-category"
)

type CategorySteps struct{}

func NewCategorySteps() *CategorySteps {
	return &CategorySteps{}
}

// Helper function to create multipart form request for category
func createCategoryMultipartRequest(category models.Category, shopID int, imageData []byte, imageType string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add category JSON
	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("category", string(categoryJSON)); err != nil {
		return nil, "", err
	}

	// Add shop_id
	if err := writer.WriteField("shop_id", fmt.Sprintf("%d", shopID)); err != nil {
		return nil, "", err
	}

	// Add image if provided
	if imageData != nil {
		filename := "image.png"
		if imageType == invalidImageType {
			filename = "image.txt"
		}

		part, err := writer.CreateFormFile("image", filename)
		if err != nil {
			return nil, "", err
		}

		if _, err := io.Copy(part, bytes.NewReader(imageData)); err != nil {
			return nil, "", err
		}
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	return body, contentType, nil
}

func (c *CategorySteps) setupSQLExpectations() {
	ctx := GetTestContext()

	switch ctx.scenario {
	case validCategoryScenario:
		// Mock successful category creation
		ctx.mockSQLMock.ExpectQuery("SELECT create_category").
			WillReturnRows(sqlmock.NewRows([]string{"create_category"}).AddRow(1))

	case duplicateCategoryScenario:
		// Mock duplicate key error
		pqErr := &pq.Error{
			Code:       "23505",
			Constraint: "categories_name_shop_id_unique",
		}
		ctx.mockSQLMock.ExpectQuery("SELECT create_category").
			WillReturnError(pqErr)
	}
}

func (c *CategorySteps) iHaveValidCategoryDataWithImage() error {
	ctx := GetTestContext()
	ctx.scenario = validCategoryScenario

	ctx.requestBody = models.Category{
		Name:        "Hamburguesas",
		Description: "Deliciosas hamburguesas con ingredients frescos",
	}

	// Create one test image
	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iHaveValidCategoryDataWithoutImage() error {
	ctx := GetTestContext()
	ctx.scenario = validCategoryScenario

	ctx.requestBody = models.Category{
		Name:        "Bebidas",
		Description: "Refrescos y jugos naturales",
	}

	ctx.productImages = nil // No image
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithEmptyName() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "",
		Description: "Test Description",
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithEmptyDescription() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Test Category",
		Description: "",
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithInvalidShopID() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Test Category",
		Description: "Test Description",
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "0" // Invalid

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithOversizedImage() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Test Category",
		Description: "Test Description",
	}

	// Create oversized image (> 3MB)
	oversizedImage := make([]byte, 3*1024*1024+1024) // 3MB + 1KB
	for i := range oversizedImage {
		oversizedImage[i] = byte(i % 256)
	}

	ctx.productImages = [][]byte{oversizedImage}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithInvalidImageType() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Test Category",
		Description: "Test Description",
	}

	// Create a text file instead of image
	ctx.productImages = [][]byte{[]byte("This is not an image")}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"
	ctx.invalidImageType = true

	return nil
}

func (c *CategorySteps) iHaveCategoryDataWithDuplicateName() error {
	ctx := GetTestContext()
	ctx.scenario = duplicateCategoryScenario

	ctx.requestBody = models.Category{
		Name:        "Hamburguesas", // Already exists
		Description: "Otra descripción",
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (c *CategorySteps) iSendACreateCategoryRequest() error {
	ctx := GetTestContext()

	if err := c.setupTestApp(ctx); err != nil {
		return err
	}

	c.setupSQLExpectations()

	body, contentType, err := c.createRequestBody(ctx)
	if err != nil {
		return err
	}

	resp, err := c.executeHTTPRequest(ctx, body, contentType)
	if err != nil {
		return err
	}

	ctx.response = resp
	return c.parseResponse(ctx, resp)
}

func (c *CategorySteps) setupTestApp(ctx *TestContext) error {
	if ctx.app == nil {
		return ctx.SetupCategoryTestApp()
	}
	return nil
}

func (c *CategorySteps) createRequestBody(ctx *TestContext) (*bytes.Buffer, string, error) {
	imageType := validImageType
	if ctx.invalidImageType {
		imageType = invalidImageType
	}

	category, ok := ctx.requestBody.(models.Category)
	if !ok {
		return nil, "", fmt.Errorf("expected models.Category in requestBody, got: %T", ctx.requestBody)
	}

	shopID := 0
	if ctx.pathParams != nil {
		if shopIDStr, ok := ctx.pathParams["shop_id"]; ok {
			_, _ = fmt.Sscanf(shopIDStr, "%d", &shopID)
		}
	}

	// Get image data
	var imageData []byte
	if len(ctx.productImages) > 0 {
		imageData = ctx.productImages[0]
	}

	return createCategoryMultipartRequest(
		category,
		shopID,
		imageData,
		imageType,
	)
}

func (c *CategorySteps) executeHTTPRequest(ctx *TestContext, body *bytes.Buffer, contentType string) (*http.Response, error) {
	url := ctx.server.URL + "/categories"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	return client.Do(req)
}

func (c *CategorySteps) parseResponse(ctx *TestContext, resp *http.Response) error {
	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResponse map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			ctx.errorMessage = errorResponse["error"]
		}
	} else {
		var category models.Category
		if err := json.NewDecoder(resp.Body).Decode(&category); err == nil {
			ctx.responseBody = &category
		}
	}

	return nil
}

func (c *CategorySteps) theCategoryShouldBeCreatedSuccessfully() error {
	ctx := GetTestContext()
	createdCategory, ok := ctx.responseBody.(*models.Category)
	if !ok || createdCategory == nil {
		return fmt.Errorf("expected category to be created, got: %T", ctx.responseBody)
	}
	if createdCategory.Name == "" {
		return fmt.Errorf("expected category name to be set")
	}
	return nil
}

// RegisterSteps registers all category step definitions
func (c *CategorySteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Success scenarios
	sc.Step(`^I have valid category data with image$`, c.iHaveValidCategoryDataWithImage)
	sc.Step(`^I have valid category data without image$`, c.iHaveValidCategoryDataWithoutImage)

	// HTTP validation scenarios
	sc.Step(`^I have category data with empty name$`, c.iHaveCategoryDataWithEmptyName)
	sc.Step(`^I have category data with empty description$`, c.iHaveCategoryDataWithEmptyDescription)
	sc.Step(`^I have category data with invalid shop_id$`, c.iHaveCategoryDataWithInvalidShopID)
	sc.Step(`^I have category data with oversized image$`, c.iHaveCategoryDataWithOversizedImage)
	sc.Step(`^I have category data with invalid image type$`, c.iHaveCategoryDataWithInvalidImageType)

	// Database validation scenarios
	sc.Step(`^I have category data with duplicate name$`, c.iHaveCategoryDataWithDuplicateName)

	// Action steps
	sc.Step(`^I send a create category request$`, c.iSendACreateCategoryRequest)

	// Assertion steps
	sc.Step(`^the category should be created successfully$`, c.theCategoryShouldBeCreatedSuccessfully)
}
