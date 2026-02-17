package requests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func createMultipartRequest(t *testing.T, shopJSON string, files map[string][]byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add shop JSON
	if shopJSON != "" {
		_ = writer.WriteField("shop", shopJSON)
	}

	// Add files
	for fieldName, content := range files {
		part, err := writer.CreateFormFile(fieldName, fieldName+".jpg")
		assert.NoError(t, err)
		_, err = part.Write(content)
		assert.NoError(t, err)
	}

	err := writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/shops/1", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse the multipart form
	err = req.ParseMultipartForm(7 << 20)
	assert.NoError(t, err)

	return req
}

func validShopJSON() string {
	return `{
		"name": "Test Shop",
		"slug": "test-shop",
		"email": "test@shop.com",
		"phone": "+54 11 1234-5678",
		"instagram": "@testshop"
	}`
}

// Create a minimal valid JPEG header (not a real image, but passes MIME detection)
func createJPEGBytes() []byte {
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
}

// Create a minimal valid PNG header
func createPNGBytes() []byte {
	return []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
}

// =============================================================================
// NewShopUpdateRequest Tests
// =============================================================================

func TestNewShopUpdateRequest(t *testing.T) {
	t.Run("when valid shop JSON then returns request", func(t *testing.T) {
		// Arrange
		req := createMultipartRequest(t, validShopJSON(), nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Shop", result.Shop.Name)
		assert.Equal(t, "test-shop", result.Shop.Slug)
	})

	t.Run("when shop JSON is empty then returns error", func(t *testing.T) {
		// Arrange
		req := createMultipartRequest(t, "", nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_json_required", badReqErr.Message)
	})

	t.Run("when shop JSON is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		req := createMultipartRequest(t, "   ", nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_json_required", badReqErr.Message)
	})

	t.Run("when shop JSON is invalid then returns error", func(t *testing.T) {
		// Arrange
		req := createMultipartRequest(t, "not valid json", nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_shop_json_format", badReqErr.Message)
	})

	t.Run("when logo file provided then includes in request", func(t *testing.T) {
		// Arrange
		files := map[string][]byte{
			"logo": createJPEGBytes(),
		}
		req := createMultipartRequest(t, validShopJSON(), files)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.NewLogo)
	})

	t.Run("when cover file provided then includes in request", func(t *testing.T) {
		// Arrange
		files := map[string][]byte{
			"cover": createPNGBytes(),
		}
		req := createMultipartRequest(t, validShopJSON(), files)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.NewCover)
	})

	t.Run("when both logo and cover provided then includes both", func(t *testing.T) {
		// Arrange
		files := map[string][]byte{
			"logo":  createJPEGBytes(),
			"cover": createPNGBytes(),
		}
		req := createMultipartRequest(t, validShopJSON(), files)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.NewLogo)
		assert.NotNil(t, result.NewCover)
	})

	t.Run("when no files provided then NewLogo and NewCover are nil", func(t *testing.T) {
		// Arrange
		req := createMultipartRequest(t, validShopJSON(), nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.NewLogo)
		assert.Nil(t, result.NewCover)
	})
}

// =============================================================================
// Validate Tests - Basic Fields
// =============================================================================

func TestShopUpdateRequest_Validate_BasicFields(t *testing.T) {
	t.Run("when name is empty then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "",
				Slug: "test-shop",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_name_is_required", badReqErr.Message)
	})

	t.Run("when name is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "   ",
				Slug: "test-shop",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_name_is_required", badReqErr.Message)
	})

	t.Run("when slug is empty then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_slug_is_required", badReqErr.Message)
	})

	t.Run("when slug is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "   ",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "shop_slug_is_required", badReqErr.Message)
	})

	t.Run("when name and slug are valid then returns no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// Validate Tests - Existing Images
// =============================================================================

func TestShopUpdateRequest_Validate_ExistingImages(t *testing.T) {
	t.Run("when existing image has no ID then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 0, URL: "https://example.com/logo.jpg", Type: "logo"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_valid_id", badReqErr.Message)
	})

	t.Run("when existing image has negative ID then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: -1, URL: "https://example.com/logo.jpg", Type: "logo"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_valid_id", badReqErr.Message)
	})

	t.Run("when existing image has empty URL then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "", Type: "logo"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_url", badReqErr.Message)
	})

	t.Run("when existing image has whitespace URL then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "   ", Type: "logo"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "existing_image_must_have_url", badReqErr.Message)
	})

	t.Run("when existing image has invalid type then returns error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "https://example.com/image.jpg", Type: "banner"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "image_type_must_be_logo_or_cover", badReqErr.Message)
	})

	t.Run("when existing image has type logo then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when existing image has type cover then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "https://example.com/cover.jpg", Type: "cover"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when multiple valid existing images then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
				Images: []*ShopImageRequest{
					{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
					{ID: 2, URL: "https://example.com/cover.jpg", Type: "cover"},
				},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when no existing images then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name:   "Test Shop",
				Slug:   "test-shop",
				Images: nil,
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when empty images array then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name:   "Test Shop",
				Slug:   "test-shop",
				Images: []*ShopImageRequest{},
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// ToLogoBuffer Tests
// =============================================================================

func TestShopUpdateRequest_ToLogoBuffer(t *testing.T) {
	t.Run("when no new logo then returns nil", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop:    ShopRequest{Name: "Test", Slug: "test"},
			NewLogo: nil,
		}

		// Act
		buffer, err := request.ToLogoBuffer()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, buffer)
	})
}

// =============================================================================
// ToCoverBuffer Tests
// =============================================================================

func TestShopUpdateRequest_ToCoverBuffer(t *testing.T) {
	t.Run("when no new cover then returns nil", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop:     ShopRequest{Name: "Test", Slug: "test"},
			NewCover: nil,
		}

		// Act
		buffer, err := request.ToCoverBuffer()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, buffer)
	})
}

// =============================================================================
// isValidImageType Tests (via validateImage)
// =============================================================================

func TestShopUpdateRequest_ValidateImage(t *testing.T) {
	// Note: These tests use validateImage indirectly through Validate()
	// Direct tests would require creating actual file headers with content

	t.Run("when shop has valid basic fields and no new images then no error", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Test Shop",
				Slug: "test-shop",
			},
			NewLogo:  nil,
			NewCover: nil,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestShopUpdateRequest_EdgeCases(t *testing.T) {
	t.Run("when shop JSON has extra fields then ignores them", func(t *testing.T) {
		// Arrange
		shopJSON := `{
			"name": "Test Shop",
			"slug": "test-shop",
			"unknown_field": "should be ignored"
		}`
		req := createMultipartRequest(t, shopJSON, nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Shop", result.Shop.Name)
	})

	t.Run("when shop JSON has nested payment methods then parses them", func(t *testing.T) {
		// Arrange
		shopJSON := `{
			"name": "Test Shop",
			"slug": "test-shop",
			"payment_methods": [
				{"id": 1, "name": "Cash", "is_active": true}
			]
		}`
		req := createMultipartRequest(t, shopJSON, nil)

		// Act
		result, err := NewShopUpdateRequest(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Shop.PaymentMethods, 1)
		assert.Equal(t, "Cash", result.Shop.PaymentMethods[0].Name)
	})

	t.Run("when shop has long name then accepts it", func(t *testing.T) {
		// Arrange
		longName := strings.Repeat("a", 500)
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: longName,
				Slug: "test-shop",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when shop has unicode characters then accepts them", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Tienda de Ropa",
				Slug: "tienda-de-ropa",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when shop has special characters in name then accepts them", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				Name: "Shop & Co. - The Best!",
				Slug: "shop-and-co",
			},
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// ShopRequest.ToModel Tests
// =============================================================================

func TestShopRequest_ToModel(t *testing.T) {
	t.Run("when basic fields then maps all", func(t *testing.T) {
		// Arrange
		primaryColor := "#FF5733"
		request := &ShopRequest{
			ID:           1,
			Name:         "Test Shop",
			Slug:         "test-shop",
			Email:        "shop@example.com",
			Phone:        "+54 11 1234-5678",
			Instagram:    "@testshop",
			PrimaryColor: &primaryColor,
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Equal(t, 1, shop.ID)
		assert.Equal(t, "Test Shop", shop.Name)
		assert.Equal(t, "test-shop", shop.Slug)
		assert.Equal(t, "shop@example.com", shop.Email)
		assert.Equal(t, "+54 11 1234-5678", shop.Phone)
		assert.Equal(t, "@testshop", shop.Instagram)
		assert.NotNil(t, shop.PrimaryColor)
		assert.Equal(t, "#FF5733", *shop.PrimaryColor)
	})

	t.Run("when images then maps images", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			Images: []*ShopImageRequest{
				{ID: 1, URL: "https://example.com/logo.jpg", Type: "logo"},
				{ID: 2, URL: "https://example.com/cover.jpg", Type: "cover"},
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Len(t, shop.Images, 2)
		assert.Equal(t, 1, shop.Images[0].ID)
		assert.Equal(t, "https://example.com/logo.jpg", shop.Images[0].URL)
		assert.Equal(t, "logo", shop.Images[0].Type)
	})

	t.Run("when address then maps address", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			Address: &ShopAddressRequest{
				ID:      5,
				Name:    "Av. Corrientes",
				PlaceID: "ChIJxyz",
				Lat:     -34.6037,
				Lng:     -58.3816,
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.NotNil(t, shop.Address)
		assert.Equal(t, 5, shop.Address.ID)
		assert.Equal(t, "Av. Corrientes", shop.Address.Name)
		assert.Equal(t, "ChIJxyz", shop.Address.PlaceID)
		assert.Equal(t, -34.6037, shop.Address.Lat)
	})

	t.Run("when payment methods with configs then maps all", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			PaymentMethods: []*ShopPaymentMethodRequest{
				{
					ID:       1,
					Name:     "Transfer",
					Code:     "transfer",
					IsActive: true,
					TransferConfig: &ShopTransferConfigRequest{
						ID:        1,
						CBU:       "0123456789012345678901",
						CUIL:      "20-12345678-9",
						Alias:     "mi.alias",
						OwnerName: "Owner",
					},
				},
				{
					ID:       2,
					Name:     "MercadoPago",
					Code:     "mercadopago",
					IsActive: true,
					MercadoPagoConfig: &ShopMercadoPagoConfigRequest{
						ID:          1,
						AccessToken: "TOKEN",
						PublicKey:   "KEY",
						UserID:      "UID",
					},
				},
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Len(t, shop.PaymentMethods, 2)
		assert.NotNil(t, shop.PaymentMethods[0].TransferConfig)
		assert.Equal(t, "0123456789012345678901", shop.PaymentMethods[0].TransferConfig.CBU)
		assert.Nil(t, shop.PaymentMethods[0].MercadoPagoConfig)
		assert.NotNil(t, shop.PaymentMethods[1].MercadoPagoConfig)
		assert.Equal(t, "TOKEN", shop.PaymentMethods[1].MercadoPagoConfig.AccessToken)
	})

	t.Run("when delivery methods with zones and pickup then maps all", func(t *testing.T) {
		// Arrange
		fixedPrice := 500.0
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			DeliveryMethods: []*ShopDeliveryMethodRequest{
				{
					ID:       1,
					Name:     "Delivery",
					Code:     "delivery",
					IsActive: true,
					DeliveryConfig: &ShopDeliveryConfigRequest{
						ID:         1,
						FixedPrice: &fixedPrice,
					},
					DeliveryZones: []*ShopDeliveryZoneRequest{
						{ID: 1, Name: "Norte", Price: 300},
						{ID: 2, Name: "Sur", Price: 400},
					},
				},
				{
					ID:       2,
					Name:     "Pickup",
					Code:     "pickup",
					IsActive: true,
					PickupConfig: &ShopPickupConfigRequest{
						ID:           1,
						Address:      "Calle 123",
						City:         "CABA",
						Province:     "Buenos Aires",
						PostalCode:   "C1000",
						Instructions: "Ring bell",
					},
				},
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Len(t, shop.DeliveryMethods, 2)
		assert.NotNil(t, shop.DeliveryMethods[0].DeliveryConfig)
		assert.Equal(t, 500.0, *shop.DeliveryMethods[0].DeliveryConfig.FixedPrice)
		assert.Len(t, shop.DeliveryMethods[0].DeliveryZones, 2)
		assert.Equal(t, "Norte", shop.DeliveryMethods[0].DeliveryZones[0].Name)
		assert.Nil(t, shop.DeliveryMethods[0].PickupConfig)
		assert.NotNil(t, shop.DeliveryMethods[1].PickupConfig)
		assert.Equal(t, "Calle 123", shop.DeliveryMethods[1].PickupConfig.Address)
	})

	t.Run("when operating schedules then maps all", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			OperatingSchedules: []*ShopOperatingScheduleRequest{
				{ID: 1, DayOfWeek: 1, OpenTime: "09:00", CloseTime: "18:00"},
				{ID: 2, DayOfWeek: 5, OpenTime: "09:00", CloseTime: "20:00"},
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Len(t, shop.OperatingSchedules, 2)
		assert.Equal(t, 1, int(shop.OperatingSchedules[0].DayOfWeek))
		assert.Equal(t, "09:00", shop.OperatingSchedules[0].OpenTime)
	})

	t.Run("when timezone then maps timezone", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{
			Name: "Test",
			Slug: "test",
			Timezone: &ShopTimezoneRequest{
				ID:   1,
				Name: "Argentina",
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.NotNil(t, shop.Timezone)
		assert.Equal(t, 1, shop.Timezone.ID)
		assert.Equal(t, "Argentina", shop.Timezone.Name)
	})

	t.Run("when no optional fields then all are nil", func(t *testing.T) {
		// Arrange
		request := &ShopRequest{Name: "Minimal", Slug: "minimal"}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Nil(t, shop.PrimaryColor)
		assert.Nil(t, shop.Images)
		assert.Nil(t, shop.Address)
		assert.Nil(t, shop.PaymentMethods)
		assert.Nil(t, shop.DeliveryMethods)
		assert.Nil(t, shop.OperatingSchedules)
		assert.Nil(t, shop.Timezone)
	})
}

// =============================================================================
// ShopUpdateRequest.ToModel Tests
// =============================================================================

func TestShopUpdateRequest_ToModel(t *testing.T) {
	t.Run("when update request then delegates to shop ToModel", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{
				ID:   1,
				Name: "Updated Shop",
				Slug: "updated-shop",
			},
		}

		// Act
		shop := request.ToModel()

		// Assert
		assert.Equal(t, 1, shop.ID)
		assert.Equal(t, "Updated Shop", shop.Name)
		assert.Equal(t, "updated-shop", shop.Slug)
	})

	t.Run("when logo and cover buffers with nil files then returns nil", func(t *testing.T) {
		// Arrange
		request := &ShopUpdateRequest{
			Shop: ShopRequest{Name: "Test", Slug: "test"},
		}

		// Act
		logo, logoErr := request.ToLogoBuffer()
		cover, coverErr := request.ToCoverBuffer()

		// Assert
		assert.NoError(t, logoErr)
		assert.Nil(t, logo)
		assert.NoError(t, coverErr)
		assert.Nil(t, cover)
	})
}
