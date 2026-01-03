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
	scenarioShopUpdateNoImages       = "shop-update-no-images"
	scenarioShopUpdateWithLogo       = "shop-update-with-logo"
	scenarioShopUpdateWithCover      = "shop-update-with-cover"
	scenarioShopUpdateWithBothImages = "shop-update-with-both-images"
	scenarioShopUpdateNoName         = "shop-update-no-name"
	scenarioShopUpdateNoSlug         = "shop-update-no-slug"
	scenarioShopUpdateInvalidImage   = "shop-update-invalid-image"
	scenarioShopUpdateInvalidCBU     = "shop-update-invalid-cbu"
	scenarioShopUpdateNotFound       = "shop-update-not-found"
)

type UpdateShopSteps struct{}

func NewUpdateShopSteps() *UpdateShopSteps {
	return &UpdateShopSteps{}
}

// ===== Given Steps =====

func (u *UpdateShopSteps) aShopWithIDExistsForUpdate(shopID int) error {
	ctx := GetTestContext()
	if ctx.pathParams == nil {
		ctx.pathParams = make(map[string]string)
	}
	ctx.pathParams["shop_id"] = fmt.Sprintf("%d", shopID)
	return nil
}

func (u *UpdateShopSteps) iHaveValidShopUpdateDataWithoutNewImages() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateNoImages
	return nil
}

func (u *UpdateShopSteps) iHaveValidShopUpdateDataWithANewLogoImage() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateWithLogo
	return nil
}

func (u *UpdateShopSteps) iHaveValidShopUpdateDataWithANewCoverImage() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateWithCover
	return nil
}

func (u *UpdateShopSteps) iHaveValidShopUpdateDataWithBothNewImages() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateWithBothImages
	return nil
}

func (u *UpdateShopSteps) iHaveShopUpdateDataWithoutName() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateNoName
	return nil
}

func (u *UpdateShopSteps) iHaveShopUpdateDataWithoutSlug() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateNoSlug
	return nil
}

func (u *UpdateShopSteps) iHaveShopUpdateDataWithAnInvalidImageType() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateInvalidImage
	ctx.invalidImageType = true
	return nil
}

func (u *UpdateShopSteps) iHaveShopUpdateDataWithInvalidTransferCBU() error {
	ctx := GetTestContext()
	ctx.scenario = scenarioShopUpdateInvalidCBU
	return nil
}

// ===== When Steps =====

func (u *UpdateShopSteps) iSendAnUpdateShopRequestForShop(shopID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	u.setupUpdateShopSQLExpectations(shopID)

	// Build multipart request
	body, contentType, err := u.createShopUpdateRequest()
	if err != nil {
		return err
	}

	// Build URL and make request
	url := fmt.Sprintf("%s/shops/%d", ctx.server.URL, shopID)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)

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
	u.parseUpdateShopResponse(ctx, resp)

	return nil
}

func (u *UpdateShopSteps) iSendAnUnauthenticatedUpdateShopRequestForShop(shopID int) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupShopTestApp(); err != nil {
			return err
		}
	}

	// Build multipart request with valid data
	ctx.scenario = scenarioShopUpdateNoImages
	body, contentType, err := u.createShopUpdateRequest()
	if err != nil {
		return err
	}

	// Build URL and make request WITHOUT auth header
	url := fmt.Sprintf("%s/shops/%d", ctx.server.URL, shopID)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)

	// No authorization header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	u.parseUpdateShopResponse(ctx, resp)

	return nil
}

// ===== Helper Functions =====

func (u *UpdateShopSteps) createShopUpdateRequest() (*bytes.Buffer, string, error) {
	shop := u.buildShopForScenario()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add shop JSON
	shopJSON, err := json.Marshal(shop)
	if err != nil {
		return nil, "", err
	}
	if err := writer.WriteField("shop", string(shopJSON)); err != nil {
		return nil, "", err
	}

	// Add images based on scenario
	if err := u.addImagesToRequest(writer); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func (u *UpdateShopSteps) addImagesToRequest(writer *multipart.Writer) error {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioShopUpdateWithLogo:
		return u.addImageToRequest(writer, "logo")
	case scenarioShopUpdateWithCover:
		return u.addImageToRequest(writer, "cover")
	case scenarioShopUpdateWithBothImages:
		if err := u.addImageToRequest(writer, "logo"); err != nil {
			return err
		}
		return u.addImageToRequest(writer, "cover")
	case scenarioShopUpdateInvalidImage:
		return u.addInvalidImageToRequest(writer, "logo")
	}
	return nil
}

func (u *UpdateShopSteps) buildShopForScenario() models.Shop {
	ctx := GetTestContext()

	shop := models.Shop{
		Name:      "Updated Shop",
		Slug:      "updated-shop",
		Email:     "updated@shop.com",
		Phone:     "+54111234567",
		Instagram: "@updatedshop",
		Images: []*models.Image{
			{ID: 1, URL: "https://cloudinary.com/existing_logo.jpg", Type: "logo"},
		},
		PaymentMethods: []*models.PaymentMethod{
			{ID: 1, Name: "Transfer", Code: "transfer", IsActive: true},
		},
		DeliveryMethods: []*models.DeliveryMethod{
			{ID: 1, Name: "Delivery", Code: "delivery", IsActive: true},
		},
	}

	switch ctx.scenario {
	case scenarioShopUpdateNoName:
		shop.Name = ""
	case scenarioShopUpdateNoSlug:
		shop.Slug = ""
	case scenarioShopUpdateInvalidCBU:
		shop.PaymentMethods = []*models.PaymentMethod{
			{
				ID:       1,
				Name:     "Transfer",
				Code:     "transfer",
				IsActive: true,
				TransferConfig: &models.TransferConfig{
					CBU:       "invalid_cbu", // Invalid - should be 22 characters
					CUIL:      "20-12345678-9",
					OwnerName: "John Doe",
				},
			},
		}
	}

	return shop
}

func (u *UpdateShopSteps) addImageToRequest(writer *multipart.Writer, imageType string) error {
	// Create a valid PNG image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	part, err := writer.CreateFormFile(imageType, fmt.Sprintf("%s.png", imageType))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, &buf)
	return err
}

func (u *UpdateShopSteps) addInvalidImageToRequest(writer *multipart.Writer, imageType string) error {
	// Create an invalid file (not an image)
	invalidContent := []byte("this is not a valid image file")

	part, err := writer.CreateFormFile(imageType, fmt.Sprintf("%s.txt", imageType))
	if err != nil {
		return err
	}
	_, err = part.Write(invalidContent)
	return err
}

// ===== SQL Mock Setup =====

func (u *UpdateShopSteps) setupUpdateShopSQLExpectations(shopID int) {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioShopUpdateNoImages, scenarioShopUpdateWithLogo, scenarioShopUpdateWithCover, scenarioShopUpdateWithBothImages:
		// For successful scenarios, expect the update stored procedure call
		ctx.mockSQLMock.ExpectQuery("SELECT update_shop").
			WillReturnRows(sqlmock.NewRows([]string{"deleted_refs"}).AddRow("{}"))

	case scenarioShopUpdateInvalidCBU:
		// Validation happens before DB call, but we still need to mock in case it reaches DB
		ctx.mockSQLMock.ExpectQuery("SELECT update_shop").
			WillReturnRows(sqlmock.NewRows([]string{"deleted_refs"}).AddRow("{}"))

	case scenarioShopUpdateNotFound, scenarioShopNotFound:
		// Mock not found - the stored procedure returns empty or we check for shop existence
		columns := []string{
			"id", "name", "slug", "email", "phone", "instagram", "created_at",
			"images", "address", "payment_methods", "delivery_methods", "operating_schedules",
		}
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM shops").
			WithArgs(shopID).
			WillReturnRows(sqlmock.NewRows(columns))

	case scenarioShopNotOwned:
		// This is handled by authorization check in handler, no SQL mock needed
	}
}

func (u *UpdateShopSteps) parseUpdateShopResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResponse map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			ctx.errorMessage = errorResponse["error"]
		}
	}
	// 204 No Content has no body to parse
}

// ===== Register Steps =====

func (u *UpdateShopSteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^a shop with ID (\d+) exists for update$`, u.aShopWithIDExistsForUpdate)
	sc.Step(`^I have valid shop update data without new images$`, u.iHaveValidShopUpdateDataWithoutNewImages)
	sc.Step(`^I have valid shop update data with a new logo image$`, u.iHaveValidShopUpdateDataWithANewLogoImage)
	sc.Step(`^I have valid shop update data with a new cover image$`, u.iHaveValidShopUpdateDataWithANewCoverImage)
	sc.Step(`^I have valid shop update data with both new images$`, u.iHaveValidShopUpdateDataWithBothNewImages)
	sc.Step(`^I have shop update data without name$`, u.iHaveShopUpdateDataWithoutName)
	sc.Step(`^I have shop update data without slug$`, u.iHaveShopUpdateDataWithoutSlug)
	sc.Step(`^I have shop update data with an invalid image type$`, u.iHaveShopUpdateDataWithAnInvalidImageType)
	sc.Step(`^I have shop update data with invalid transfer CBU$`, u.iHaveShopUpdateDataWithInvalidTransferCBU)

	// When steps
	sc.Step(`^I send an update shop request for shop (\d+)$`, u.iSendAnUpdateShopRequestForShop)
	sc.Step(`^I send an unauthenticated update shop request for shop (\d+)$`, u.iSendAnUnauthenticatedUpdateShopRequestForShop)
}
