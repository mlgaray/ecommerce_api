package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

const (
	invalidImageType     = "invalid"
	validProductScenario = "valid-product"
	testImageSize        = 100 // Default test image size
)

type ProductSteps struct{}

func NewProductSteps() *ProductSteps {
	return &ProductSteps{}
}

// Helper function to create a valid PNG image in memory
func createTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, testImageSize, testImageSize))
	// Fill with a simple color
	for y := 0; y < testImageSize; y++ {
		for x := 0; x < testImageSize; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}

// Helper function to create multipart form request
func createMultipartRequest(product models.Product, shopID int, images [][]byte, imageType string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add product JSON
	productJSON, err := json.Marshal(product)
	if err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("product", string(productJSON)); err != nil {
		return nil, "", err
	}

	// Add shop_id
	if err := writer.WriteField("shop_id", fmt.Sprintf("%d", shopID)); err != nil {
		return nil, "", err
	}

	// Add images
	for i, imageData := range images {
		filename := fmt.Sprintf("image%d.png", i)
		if imageType == invalidImageType {
			filename = fmt.Sprintf("image%d.txt", i)
		}

		part, err := writer.CreateFormFile(fmt.Sprintf("images[%d]", i), filename)
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

func (p *ProductSteps) setupSQLExpectations() {
	ctx := GetTestContext()

	if ctx.scenario == validProductScenario {
		// Mock successful product creation
		ctx.mockSQLMock.ExpectBegin()

		// Expect INSERT for product
		ctx.mockSQLMock.ExpectQuery("INSERT INTO products").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect INSERT for product images
		ctx.mockSQLMock.ExpectExec("INSERT INTO product_images").
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx.mockSQLMock.ExpectCommit()
	}
}

func (p *ProductSteps) iHaveValidProductDataWithImages() error {
	ctx := GetTestContext()
	ctx.scenario = validProductScenario

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	// Create one test image (100x100 PNG)
	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithoutImages() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{} // No images
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithEmptyName() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithEmptyDescription() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithNegativePrice() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        -10.00,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithNegativeStock() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        -5,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithoutCategory() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     nil,
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithInvalidShopID() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	ctx.productImages = [][]byte{createTestImage()}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "0" // Invalid shop ID

	return nil
}

func (p *ProductSteps) iHaveProductDataWithOversizedImage() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
	}

	// Create oversized image (> 3MB but < 13MB)
	// Generate a buffer of 3MB + 1KB to ensure it exceeds the 3MB limit
	oversizedImage := make([]byte, 3*1024*1024+1024) // 3MB + 1KB
	for i := range oversizedImage {
		oversizedImage[i] = byte(i % 256) // Fill with pattern data
	}

	ctx.productImages = [][]byte{oversizedImage}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (p *ProductSteps) iHaveProductDataWithInvalidImageType() error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        99.99,
		Stock:        10,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
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

func (p *ProductSteps) iSendACreateProductRequest() error {
	ctx := GetTestContext()

	if err := p.setupTestApp(ctx); err != nil {
		return err
	}

	p.setupSQLExpectations()

	body, contentType, err := p.createRequestBody(ctx)
	if err != nil {
		return err
	}

	resp, err := p.executeHTTPRequest(ctx, body, contentType)
	if err != nil {
		return err
	}

	ctx.response = resp
	return p.parseResponse(ctx, resp)
}

func (p *ProductSteps) setupTestApp(ctx *TestContext) error {
	if ctx.app == nil {
		return ctx.SetupProductTestApp()
	}
	return nil
}

func (p *ProductSteps) createRequestBody(ctx *TestContext) (*bytes.Buffer, string, error) {
	imageType := "png"
	if ctx.invalidImageType {
		imageType = invalidImageType
	}

	product, ok := ctx.requestBody.(models.Product)
	if !ok {
		return nil, "", fmt.Errorf("expected models.Product in requestBody, got: %T", ctx.requestBody)
	}

	shopID := 0
	if ctx.pathParams != nil {
		if shopIDStr, ok := ctx.pathParams["shop_id"]; ok {
			_, _ = fmt.Sscanf(shopIDStr, "%d", &shopID)
		}
	}

	return createMultipartRequest(
		product,
		shopID,
		ctx.productImages,
		imageType,
	)
}

func (p *ProductSteps) executeHTTPRequest(ctx *TestContext, body *bytes.Buffer, contentType string) (*http.Response, error) {
	url := ctx.server.URL + "/products"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	return client.Do(req)
}

func (p *ProductSteps) parseResponse(ctx *TestContext, resp *http.Response) error {
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
		var product models.Product
		if err := json.NewDecoder(resp.Body).Decode(&product); err == nil {
			ctx.responseBody = &product
		}
	}

	return nil
}

func (p *ProductSteps) theProductShouldBeCreatedSuccessfully() error {
	ctx := GetTestContext()
	createdProduct, ok := ctx.responseBody.(*models.Product)
	if !ok || createdProduct == nil {
		return fmt.Errorf("expected product to be created, got: %T", ctx.responseBody)
	}
	if createdProduct.Name == "" {
		return fmt.Errorf("expected product name to be set")
	}
	return nil
}

// RegisterSteps registers all step definitions
func (p *ProductSteps) RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I have valid product data with images$`, p.iHaveValidProductDataWithImages)
	sc.Step(`^I have product data without images$`, p.iHaveProductDataWithoutImages)
	sc.Step(`^I have product data with empty name$`, p.iHaveProductDataWithEmptyName)
	sc.Step(`^I have product data with empty description$`, p.iHaveProductDataWithEmptyDescription)
	sc.Step(`^I have product data with negative price$`, p.iHaveProductDataWithNegativePrice)
	sc.Step(`^I have product data with negative stock$`, p.iHaveProductDataWithNegativeStock)
	sc.Step(`^I have product data without category$`, p.iHaveProductDataWithoutCategory)
	sc.Step(`^I have product data with invalid shop_id$`, p.iHaveProductDataWithInvalidShopID)
	sc.Step(`^I have product data with oversized image$`, p.iHaveProductDataWithOversizedImage)
	sc.Step(`^I have product data with invalid image type$`, p.iHaveProductDataWithInvalidImageType)
	sc.Step(`^I send a create product request$`, p.iSendACreateProductRequest)
	sc.Step(`^the product should be created successfully$`, p.theProductShouldBeCreatedSuccessfully)
}
