package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	validUpdateScenario = "valid-update"
)

type UpdateProductSteps struct{}

func NewUpdateProductSteps() *UpdateProductSteps {
	return &UpdateProductSteps{}
}

func (u *UpdateProductSteps) setupSQLExpectations() {
	ctx := GetTestContext()

	if ctx.scenario == validUpdateScenario {
		// Mock successful product update via stored procedure
		// The SP returns an array of deleted storage_refs for Cloudinary cleanup
		rows := sqlmock.NewRows([]string{"update_product"}).
			AddRow("{}") // Empty array - no images deleted
		ctx.mockSQLMock.ExpectQuery("SELECT update_product").
			WillReturnRows(rows)

		// Mock getByID query that handler calls after update to return updated product
		productID := 1
		if id, ok := ctx.pathParams["product_id"]; ok {
			fmt.Sscanf(id, "%d", &productID)
		}

		getByIDColumns := []string{
			"id", "name", "description", "price", "is_stockeable", "stock", "minimum_stock",
			"is_active", "is_highlighted", "is_promotional", "promotional_price",
			"created_at", "category_id", "category_name", "category_description",
			"images", "variants",
		}
		imagesJSON := `[{"id": 1, "url": "https://existing.com/image1.jpg"}]`
		getByIDRows := sqlmock.NewRows(getByIDColumns).
			AddRow(productID, "Updated Product", "Updated Description", 149.99, false, 20, 5,
				true, false, false, 0.0, time.Now(), 1, "Category 1", "", imagesJSON, "[]")
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM products").
			WithArgs(productID).
			WillReturnRows(getByIDRows)
	}
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndValidUpdateData(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = validUpdateScenario

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndKeepExistingImages(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = validUpdateScenario

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
			{ID: 2, URL: "https://existing.com/image2.jpg"},
		},
	}

	ctx.productImages = [][]byte{} // No new images
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndAddNewImages(productID int) error {
	ctx := GetTestContext()
	ctx.scenario = validUpdateScenario

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	ctx.productImages = [][]byte{createTestImage()} // Add new image
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndNoImages(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images:       []*models.Image{}, // No existing images
	}

	ctx.productImages = [][]byte{} // No new images
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndEmptyName(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "", // Empty name
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndEmptyDescription(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "", // Empty description
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndNoCategory(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     nil, // No category
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndOversizedNewImage(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	// Create oversized image (> 3MB)
	oversizedImage := make([]byte, 3*1024*1024+1024)
	for i := range oversizedImage {
		oversizedImage[i] = byte(i % 256)
	}

	ctx.productImages = [][]byte{oversizedImage}
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndInvalidNewImageType(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	ctx.productImages = [][]byte{[]byte("This is not an image")}
	ctx.invalidImageType = true

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAnInvalidProductID(invalidID string) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = invalidID // Invalid format
	ctx.pathParams["shop_id"] = "1"

	return nil
}

// Business validation scenarios
func (u *UpdateProductSteps) iHaveAProductWithIDAndNegativePrice(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        -10.00, // Negative price
		Stock:        20,
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndNegativeStock(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		Stock:        -5, // Negative stock
		MinimumStock: 5,
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAndNegativeMinimumStock(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		IsStockeable: true, // Required for minimum_stock validation
		Stock:        20,
		MinimumStock: -5, // Negative minimum stock
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDWithMinimumStockButNoStock(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		IsStockeable: true, // Required for minimum_stock validation
		Stock:        0,    // No stock
		MinimumStock: 5,    // But has minimum stock
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDWithMinimumStockGreaterThanStock(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:           productID,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Price:        149.99,
		IsStockeable: true, // Required for minimum_stock validation
		Stock:        10,   // Stock is 10
		MinimumStock: 20,   // But minimum stock is 20
		Category:     &models.Category{ID: 1},
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDAsPromotionalWithoutPromotionalPrice(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:               productID,
		Name:             "Updated Product",
		Description:      "Updated Description",
		Price:            149.99,
		Stock:            20,
		MinimumStock:     5,
		Category:         &models.Category{ID: 1},
		IsPromotional:    true, // Is promotional
		PromotionalPrice: 0,    // But no promotional price
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iHaveAProductWithIDWithPromotionalPriceNotLowerThanPrice(productID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Product{
		ID:               productID,
		Name:             "Updated Product",
		Description:      "Updated Description",
		Price:            100.00, // Regular price
		Stock:            20,
		MinimumStock:     5,
		Category:         &models.Category{ID: 1},
		IsPromotional:    true,
		PromotionalPrice: 120.00, // Promotional price is higher!
		Images: []*models.Image{
			{ID: 1, URL: "https://existing.com/image1.jpg"},
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["product_id"] = fmt.Sprintf("%d", productID)
	ctx.pathParams["shop_id"] = "1"

	return nil
}

func (u *UpdateProductSteps) iSendAnUpdateProductRequest() error {
	ctx := GetTestContext()

	if err := u.setupTestApp(ctx); err != nil {
		return err
	}

	u.setupSQLExpectations()

	body, contentType, err := u.createRequestBody(ctx)
	if err != nil {
		return err
	}

	resp, err := u.executeHTTPRequest(ctx, body, contentType)
	if err != nil {
		return err
	}

	ctx.response = resp
	return u.parseResponse(ctx, resp)
}

func (u *UpdateProductSteps) iSendAnUpdateProductRequestWithInvalidID() error {
	// Same as regular update but with invalid ID already set in path params
	return u.iSendAnUpdateProductRequest()
}

func (u *UpdateProductSteps) setupTestApp(ctx *TestContext) error {
	if ctx.app == nil {
		return ctx.SetupProductTestApp()
	}
	return nil
}

func (u *UpdateProductSteps) createRequestBody(ctx *TestContext) (*bytes.Buffer, string, error) {
	imageType := validImageType
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

	return createMultipartUpdateRequest(
		product,
		shopID,
		ctx.productImages,
		imageType,
	)
}

func createMultipartUpdateRequest(product models.Product, shopID int, newImages [][]byte, imageType string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add product JSON (includes existing images)
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

	// Add new images (if any)
	for i, imageData := range newImages {
		filename := fmt.Sprintf("new_image%d.png", i)
		if imageType == invalidImageType {
			filename = fmt.Sprintf("new_image%d.txt", i)
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

func (u *UpdateProductSteps) executeHTTPRequest(ctx *TestContext, body *bytes.Buffer, contentType string) (*http.Response, error) {
	productID := ctx.pathParams["product_id"]
	url := ctx.server.URL + "/products/" + productID
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	// Add Authorization header with JWT token
	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	return client.Do(req)
}

func (u *UpdateProductSteps) parseResponse(ctx *TestContext, resp *http.Response) error {
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
		var successResponse responses.UpdateProductResponse
		if err := json.NewDecoder(resp.Body).Decode(&successResponse); err == nil {
			ctx.responseBody = successResponse.Product
		}
	}

	return nil
}

func (u *UpdateProductSteps) theUserShouldReceiveTheUpdatedProduct() error {
	ctx := GetTestContext()
	product, ok := ctx.responseBody.(*responses.ProductResponse)
	if !ok {
		return fmt.Errorf("expected ProductResponse, got: %T", ctx.responseBody)
	}
	if product == nil || product.ID == 0 {
		return fmt.Errorf("expected product with ID, got: %v", product)
	}
	if product.Name == "" {
		return fmt.Errorf("expected product with name, got empty name")
	}
	return nil
}

// RegisterSteps registers all step definitions
func (u *UpdateProductSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Success scenarios
	sc.Step(`^I have a product with id (\d+) and valid update data$`, u.iHaveAProductWithIDAndValidUpdateData)
	sc.Step(`^I have a product with id (\d+) and keep existing images$`, u.iHaveAProductWithIDAndKeepExistingImages)
	sc.Step(`^I have a product with id (\d+) and add new images$`, u.iHaveAProductWithIDAndAddNewImages)

	// HTTP validation scenarios
	sc.Step(`^I have a product with id (\d+) and no images$`, u.iHaveAProductWithIDAndNoImages)
	sc.Step(`^I have a product with id (\d+) and empty name$`, u.iHaveAProductWithIDAndEmptyName)
	sc.Step(`^I have a product with id (\d+) and empty description$`, u.iHaveAProductWithIDAndEmptyDescription)
	sc.Step(`^I have a product with id (\d+) and no category$`, u.iHaveAProductWithIDAndNoCategory)
	sc.Step(`^I have a product with id (\d+) and oversized new image$`, u.iHaveAProductWithIDAndOversizedNewImage)
	sc.Step(`^I have a product with id (\d+) and invalid new image type$`, u.iHaveAProductWithIDAndInvalidNewImageType)
	sc.Step(`^I have an invalid product_id "([^"]*)"$`, u.iHaveAnInvalidProductID)

	// Business validation scenarios
	sc.Step(`^I have a product with id (\d+) and negative price$`, u.iHaveAProductWithIDAndNegativePrice)
	sc.Step(`^I have a product with id (\d+) and negative stock$`, u.iHaveAProductWithIDAndNegativeStock)
	sc.Step(`^I have a product with id (\d+) and negative minimum stock$`, u.iHaveAProductWithIDAndNegativeMinimumStock)
	sc.Step(`^I have a product with id (\d+) with minimum stock but no stock$`, u.iHaveAProductWithIDWithMinimumStockButNoStock)
	sc.Step(`^I have a product with id (\d+) with minimum stock greater than stock$`, u.iHaveAProductWithIDWithMinimumStockGreaterThanStock)
	sc.Step(`^I have a product with id (\d+) as promotional without promotional price$`, u.iHaveAProductWithIDAsPromotionalWithoutPromotionalPrice)
	sc.Step(`^I have a product with id (\d+) with promotional price not lower than price$`, u.iHaveAProductWithIDWithPromotionalPriceNotLowerThanPrice)

	// Action steps
	sc.Step(`^I send an update product request$`, u.iSendAnUpdateProductRequest)
	sc.Step(`^I send an update product request with invalid id$`, u.iSendAnUpdateProductRequestWithInvalidID)
	sc.Step(`^the user should receive the updated product$`, u.theUserShouldReceiveTheUpdatedProduct)
}
