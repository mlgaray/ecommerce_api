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
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/responses"
)

const (
	scenarioUpdateCategoryValid         = "update-category-valid"
	scenarioUpdateCategoryNotFound      = "update-category-not-found"
	scenarioUpdateCategoryDuplicateName = "update-category-duplicate-name"
)

type UpdateCategorySteps struct{}

func NewUpdateCategorySteps() *UpdateCategorySteps {
	return &UpdateCategorySteps{}
}

// ===== Given Steps =====

func (u *UpdateCategorySteps) theUserIsAuthenticatedForCategoryUpdateOperations() error {
	ctx := GetTestContext()
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}
	return nil
}

func (u *UpdateCategorySteps) theUserIsNotAuthenticatedForCategoryUpdate() error {
	ctx := GetTestContext()
	ctx.authToken = "" // Clear auth token
	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndValidUpdateDataWithExistingImage(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateCategoryValid

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndValidUpdateDataWithNewImage(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateCategoryValid

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		// No existing image - will add new image
	}

	// Create a valid PNG image for upload
	ctx.productImages = [][]byte{createTestImage()}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndEmptyNameForUpdate(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "", // Empty name
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndEmptyDescriptionForUpdate(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "", // Empty description
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndNoImageForUpdate(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image:       nil, // No image
	}

	ctx.productImages = [][]byte{} // No new image either

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndExistingImageWithoutID(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  0, // Invalid ID
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndExistingImageWithoutURL(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "", // Empty URL
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndOversizedNewImageForUpdate(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		// No existing image - will provide oversized new image
	}

	// Create oversized image (> 3MB)
	ctx.productImages = [][]byte{make([]byte, 4*1024*1024)}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndInvalidImageTypeForUpdate(categoryID int) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		// No existing image - will provide invalid type
	}

	// Create a text file instead of image
	ctx.productImages = [][]byte{[]byte("This is not an image file")}
	ctx.invalidImageType = true

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveAnInvalidCategoryIDForUpdate(invalidID string) error {
	ctx := GetTestContext()

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = invalidID

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDThatDoesNotExist(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateCategoryNotFound

	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

func (u *UpdateCategorySteps) iHaveACategoryWithIDAndANameThatAlreadyExists(categoryID int) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioUpdateCategoryDuplicateName

	ctx.requestBody = models.Category{
		Name:        "Existing Category Name",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return nil
}

// ===== When Steps =====

func (u *UpdateCategorySteps) iSendAnUpdateCategoryRequest() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	u.setupUpdateCategorySQLExpectations()

	categoryID := ctx.pathParams["category_id"]
	return u.sendUpdateCategoryRequest(categoryID)
}

func (u *UpdateCategorySteps) iSendAnUpdateCategoryRequestWithInvalidID() error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	categoryID := ctx.pathParams["category_id"]
	return u.sendUpdateCategoryRequest(categoryID)
}

func (u *UpdateCategorySteps) iSendAnUnauthenticatedUpdateCategoryRequestForCategory(categoryID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupCategoryTestApp(); err != nil {
			return err
		}
	}

	// Set up request body
	ctx.requestBody = models.Category{
		Name:        "Updated Category",
		Description: "Updated Description",
		Image: &models.Image{
			ID:  5,
			URL: "https://cloudinary.com/existing_image.jpg",
		},
	}

	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["category_id"] = fmt.Sprintf("%d", categoryID)

	return u.sendUpdateCategoryRequest(fmt.Sprintf("%d", categoryID))
}

func (u *UpdateCategorySteps) sendUpdateCategoryRequest(categoryID string) error {
	ctx := GetTestContext()

	body, contentType, err := u.buildMultipartForm(ctx)
	if err != nil {
		return err
	}

	resp, err := u.executeRequest(ctx, categoryID, body, contentType)
	if err != nil {
		return err
	}

	ctx.response = resp
	u.parseResponse(ctx, resp)

	return nil
}

func (u *UpdateCategorySteps) buildMultipartForm(ctx *TestContext) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	category, ok := ctx.requestBody.(models.Category)
	if !ok {
		return nil, "", fmt.Errorf("expected models.Category, got %T", ctx.requestBody)
	}

	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return nil, "", err
	}

	if err := writer.WriteField("category", string(categoryJSON)); err != nil {
		return nil, "", err
	}

	if err := u.addImageToForm(ctx, writer); err != nil {
		return nil, "", err
	}

	writer.Close()
	return body, writer.FormDataContentType(), nil
}

func (u *UpdateCategorySteps) addImageToForm(ctx *TestContext, writer *multipart.Writer) error {
	if len(ctx.productImages) == 0 || len(ctx.productImages[0]) == 0 {
		return nil
	}

	filename := "test.png"
	if ctx.invalidImageType {
		filename = "test.txt"
	}

	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, bytes.NewReader(ctx.productImages[0]))
	return err
}

func (u *UpdateCategorySteps) executeRequest(ctx *TestContext, categoryID string, body *bytes.Buffer, contentType string) (*http.Response, error) {
	url := fmt.Sprintf("%s/categories/%s", ctx.server.URL, categoryID)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	if ctx.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.authToken)
	}

	client := &http.Client{}
	return client.Do(req)
}

func (u *UpdateCategorySteps) setupUpdateCategorySQLExpectations() {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioUpdateCategoryValid:
		// Mock successful category update via stored procedure
		rows := sqlmock.NewRows([]string{"update_category"}).
			AddRow("") // Empty string - no old image deleted
		ctx.mockSQLMock.ExpectQuery("SELECT update_category").
			WillReturnRows(rows)

	case scenarioUpdateCategoryNotFound:
		// Mock category not found error
		pqErr := &pq.Error{
			Code:    "P0003",
			Message: "Category not found or does not belong to shop",
		}
		ctx.mockSQLMock.ExpectQuery("SELECT update_category").
			WillReturnError(pqErr)

	case scenarioUpdateCategoryDuplicateName:
		// Mock duplicate name error
		pqErr := &pq.Error{
			Code:       "23505",
			Constraint: "categories_name_shop_id_unique",
		}
		ctx.mockSQLMock.ExpectQuery("SELECT update_category").
			WillReturnError(pqErr)
	}
}

func (u *UpdateCategorySteps) parseResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var response map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			if errMsg, ok := response["error"]; ok {
				ctx.errorMessage = errMsg
			}
		}
		return
	}

	var response responses.UpdateCategoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
		ctx.responseBody = response.Category
	}
}

func (u *UpdateCategorySteps) theUserShouldReceiveTheUpdatedCategory() error {
	ctx := GetTestContext()
	category, ok := ctx.responseBody.(*responses.CategoryResponse)
	if !ok {
		return fmt.Errorf("expected CategoryResponse, got: %T", ctx.responseBody)
	}
	if category == nil || category.ID == 0 {
		return fmt.Errorf("expected category with ID, got: %v", category)
	}
	if category.Name == "" {
		return fmt.Errorf("expected category with name, got empty name")
	}
	return nil
}

// ===== Register Steps =====

func (u *UpdateCategorySteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the user is authenticated for category update operations$`, u.theUserIsAuthenticatedForCategoryUpdateOperations)
	sc.Step(`^the user is not authenticated for category update$`, u.theUserIsNotAuthenticatedForCategoryUpdate)
	sc.Step(`^I have a category with id (\d+) and valid update data with existing image$`, u.iHaveACategoryWithIDAndValidUpdateDataWithExistingImage)
	sc.Step(`^I have a category with id (\d+) and valid update data with new image$`, u.iHaveACategoryWithIDAndValidUpdateDataWithNewImage)
	sc.Step(`^I have a category with id (\d+) and empty name for update$`, u.iHaveACategoryWithIDAndEmptyNameForUpdate)
	sc.Step(`^I have a category with id (\d+) and empty description for update$`, u.iHaveACategoryWithIDAndEmptyDescriptionForUpdate)
	sc.Step(`^I have a category with id (\d+) and no image for update$`, u.iHaveACategoryWithIDAndNoImageForUpdate)
	sc.Step(`^I have a category with id (\d+) and existing image without ID$`, u.iHaveACategoryWithIDAndExistingImageWithoutID)
	sc.Step(`^I have a category with id (\d+) and existing image without URL$`, u.iHaveACategoryWithIDAndExistingImageWithoutURL)
	sc.Step(`^I have a category with id (\d+) and oversized new image for update$`, u.iHaveACategoryWithIDAndOversizedNewImageForUpdate)
	sc.Step(`^I have a category with id (\d+) and invalid image type for update$`, u.iHaveACategoryWithIDAndInvalidImageTypeForUpdate)
	sc.Step(`^I have an invalid category_id "([^"]*)" for update$`, u.iHaveAnInvalidCategoryIDForUpdate)
	sc.Step(`^I have a category with id (\d+) that does not exist$`, u.iHaveACategoryWithIDThatDoesNotExist)
	sc.Step(`^I have a category with id (\d+) and a name that already exists$`, u.iHaveACategoryWithIDAndANameThatAlreadyExists)

	// When steps
	sc.Step(`^I send an update category request$`, u.iSendAnUpdateCategoryRequest)
	sc.Step(`^I send an update category request with invalid id$`, u.iSendAnUpdateCategoryRequestWithInvalidID)
	sc.Step(`^I send an unauthenticated update category request for category (\d+)$`, u.iSendAnUnauthenticatedUpdateCategoryRequestForCategory)
	sc.Step(`^the user should receive the updated category$`, u.theUserShouldReceiveTheUpdatedCategory)
}
