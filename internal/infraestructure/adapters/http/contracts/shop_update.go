package contracts

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// ShopRequest represents shop data in HTTP requests.
type ShopRequest struct {
	ID                 int                             `json:"id,omitempty"`
	Name               string                          `json:"name"`
	Slug               string                          `json:"slug"`
	Email              string                          `json:"email,omitempty"`
	Phone              string                          `json:"phone,omitempty"`
	Instagram          string                          `json:"instagram,omitempty"`
	PrimaryColor       *string                         `json:"primary_color,omitempty"`
	Images             []*ShopImageRequest             `json:"images,omitempty"`
	Address            *ShopAddressRequest             `json:"address,omitempty"`
	PaymentMethods     []*ShopPaymentMethodRequest     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*ShopDeliveryMethodRequest    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*ShopOperatingScheduleRequest `json:"operating_schedules,omitempty"`
	Timezone           *ShopTimezoneRequest            `json:"timezone,omitempty"`
}

// ShopImageRequest represents image data in shop requests.
type ShopImageRequest struct {
	ID    int    `json:"id,omitempty"`
	URL   string `json:"url,omitempty"`
	Type  string `json:"type,omitempty"` // "logo" or "cover"
	Order int    `json:"order,omitempty"`
}

// ShopAddressRequest represents address data in shop requests.
type ShopAddressRequest struct {
	ID      int     `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}

// ShopPaymentMethodRequest represents payment method data in shop requests.
type ShopPaymentMethodRequest struct {
	ID                int                           `json:"id,omitempty"`
	Name              string                        `json:"name,omitempty"`
	Code              string                        `json:"code,omitempty"`
	IsActive          bool                          `json:"is_active"`
	TransferConfig    *ShopTransferConfigRequest    `json:"transfer_config,omitempty"`
	MercadoPagoConfig *ShopMercadoPagoConfigRequest `json:"mercadopago_config,omitempty"`
}

// ShopTransferConfigRequest represents transfer configuration in payment method requests.
type ShopTransferConfigRequest struct {
	ID        int    `json:"id,omitempty"`
	CBU       string `json:"cbu,omitempty"`
	CUIL      string `json:"cuil,omitempty"`
	Alias     string `json:"alias,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
}

// ShopMercadoPagoConfigRequest represents MercadoPago configuration in payment method requests.
type ShopMercadoPagoConfigRequest struct {
	ID          int    `json:"id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	PublicKey   string `json:"public_key,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// ShopDeliveryMethodRequest represents delivery method data in shop requests.
type ShopDeliveryMethodRequest struct {
	ID             int                        `json:"id,omitempty"`
	Name           string                     `json:"name,omitempty"`
	Code           string                     `json:"code,omitempty"`
	IsActive       bool                       `json:"is_active"`
	DeliveryConfig *ShopDeliveryConfigRequest `json:"delivery_config,omitempty"`
	DeliveryZones  []*ShopDeliveryZoneRequest `json:"delivery_zones,omitempty"`
	PickupConfig   *ShopPickupConfigRequest   `json:"pickup_config,omitempty"`
}

// ShopDeliveryConfigRequest represents delivery configuration in shop requests.
type ShopDeliveryConfigRequest struct {
	ID         int      `json:"id,omitempty"`
	FixedPrice *float64 `json:"fixed_price,omitempty"`
}

// ShopDeliveryZoneRequest represents delivery zone data in shop requests.
type ShopDeliveryZoneRequest struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Price float64 `json:"price,omitempty"`
}

// ShopPickupConfigRequest represents pickup configuration in delivery method requests.
type ShopPickupConfigRequest struct {
	ID           int    `json:"id,omitempty"`
	Address      string `json:"address,omitempty"`
	City         string `json:"city,omitempty"`
	Province     string `json:"province,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Instructions string `json:"instructions,omitempty"`
}

// ShopOperatingScheduleRequest represents operating schedule data in shop requests.
type ShopOperatingScheduleRequest struct {
	ID        int    `json:"id,omitempty"`
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time,omitempty"`
	CloseTime string `json:"close_time,omitempty"`
}

// ShopTimezoneRequest represents timezone data in shop requests.
type ShopTimezoneRequest struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// ToModel converts ShopRequest to a domain model.
//
//nolint:gocyclo // Complexity is intentional for readability - linear field mappings for nested structs (images, payment methods, delivery methods, operating schedules)
func (s *ShopRequest) ToModel() *models.Shop {
	shop := &models.Shop{
		ID:           s.ID,
		Name:         s.Name,
		Slug:         s.Slug,
		Email:        s.Email,
		Phone:        s.Phone,
		Instagram:    s.Instagram,
		PrimaryColor: s.PrimaryColor,
	}

	// Convert images
	if len(s.Images) > 0 {
		shop.Images = make([]*models.Image, len(s.Images))
		for i, img := range s.Images {
			shop.Images[i] = &models.Image{
				ID:   img.ID,
				URL:  img.URL,
				Type: img.Type,
			}
		}
	}

	// Convert address
	if s.Address != nil {
		shop.Address = &models.Address{
			ID:      s.Address.ID,
			Name:    s.Address.Name,
			PlaceID: s.Address.PlaceID,
			Lat:     s.Address.Lat,
			Lng:     s.Address.Lng,
		}
	}

	// Convert payment methods
	if len(s.PaymentMethods) > 0 {
		shop.PaymentMethods = make([]*models.PaymentMethod, len(s.PaymentMethods))
		for i, pm := range s.PaymentMethods {
			paymentMethod := &models.PaymentMethod{
				ID:       pm.ID,
				Name:     pm.Name,
				Code:     models.PaymentMethodCode(pm.Code),
				IsActive: pm.IsActive,
			}
			if pm.TransferConfig != nil {
				paymentMethod.TransferConfig = &models.TransferConfig{
					ID:        pm.TransferConfig.ID,
					CBU:       pm.TransferConfig.CBU,
					CUIL:      pm.TransferConfig.CUIL,
					Alias:     pm.TransferConfig.Alias,
					OwnerName: pm.TransferConfig.OwnerName,
				}
			}
			if pm.MercadoPagoConfig != nil {
				paymentMethod.MercadoPagoConfig = &models.MercadoPagoConfig{
					ID:          pm.MercadoPagoConfig.ID,
					AccessToken: pm.MercadoPagoConfig.AccessToken,
					PublicKey:   pm.MercadoPagoConfig.PublicKey,
					UserID:      pm.MercadoPagoConfig.UserID,
				}
			}
			shop.PaymentMethods[i] = paymentMethod
		}
	}

	// Convert delivery methods
	if len(s.DeliveryMethods) > 0 {
		shop.DeliveryMethods = make([]*models.DeliveryMethod, len(s.DeliveryMethods))
		for i, dm := range s.DeliveryMethods {
			deliveryMethod := &models.DeliveryMethod{
				ID:       dm.ID,
				Name:     dm.Name,
				Code:     models.DeliveryMethodCode(dm.Code),
				IsActive: dm.IsActive,
			}
			if dm.DeliveryConfig != nil {
				deliveryMethod.DeliveryConfig = &models.DeliveryConfig{
					ID:         dm.DeliveryConfig.ID,
					FixedPrice: dm.DeliveryConfig.FixedPrice,
				}
			}
			if len(dm.DeliveryZones) > 0 {
				deliveryMethod.DeliveryZones = make([]*models.DeliveryZone, len(dm.DeliveryZones))
				for j, zone := range dm.DeliveryZones {
					deliveryMethod.DeliveryZones[j] = &models.DeliveryZone{
						ID:    zone.ID,
						Name:  zone.Name,
						Price: zone.Price,
					}
				}
			}
			if dm.PickupConfig != nil {
				deliveryMethod.PickupConfig = &models.PickupConfig{
					ID:           dm.PickupConfig.ID,
					Address:      dm.PickupConfig.Address,
					City:         dm.PickupConfig.City,
					Province:     dm.PickupConfig.Province,
					PostalCode:   dm.PickupConfig.PostalCode,
					Instructions: dm.PickupConfig.Instructions,
				}
			}
			shop.DeliveryMethods[i] = deliveryMethod
		}
	}

	// Convert operating schedules
	if len(s.OperatingSchedules) > 0 {
		shop.OperatingSchedules = make([]*models.OperatingSchedule, len(s.OperatingSchedules))
		for i, os := range s.OperatingSchedules {
			shop.OperatingSchedules[i] = &models.OperatingSchedule{
				ID:        os.ID,
				DayOfWeek: models.DayOfWeek(os.DayOfWeek),
				OpenTime:  os.OpenTime,
				CloseTime: os.CloseTime,
			}
		}
	}

	// Convert timezone
	if s.Timezone != nil {
		shop.Timezone = &models.Timezone{
			ID:   s.Timezone.ID,
			Name: s.Timezone.Name,
		}
	}

	return shop
}

// ShopUpdateRequest represents the multipart form request for updating a shop.
// Form fields:
// - "shop": JSON string with shop data (required)
// - "logo": optional new logo image file
// - "cover": optional new cover image file
type ShopUpdateRequest struct {
	Shop     ShopRequest           `json:"shop"`
	NewLogo  *multipart.FileHeader `json:"-"` // Optional new logo image
	NewCover *multipart.FileHeader `json:"-"` // Optional new cover image
}

// NewShopUpdateRequest creates a ShopUpdateRequest from an HTTP request.
//
//nolint:gocyclo // Complexity is intentional for readability - multipart form parsing
func NewShopUpdateRequest(r *http.Request) (*ShopUpdateRequest, error) {
	// Extract shop JSON from form data
	shopJSON := r.FormValue("shop")
	if strings.TrimSpace(shopJSON) == "" {
		return nil, &httpErrors.BadRequestError{Message: "shop_json_required"}
	}

	// Parse shop JSON (includes existing images with IDs)
	var shop ShopRequest
	if err := json.Unmarshal([]byte(shopJSON), &shop); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "invalid_shop_json_format"}
	}

	// Get new logo from multipart form (optional)
	var newLogo *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["logo"]; exists && len(files) > 0 {
			newLogo = files[0]
		}
	}

	// Get new cover from multipart form (optional)
	var newCover *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["cover"]; exists && len(files) > 0 {
			newCover = files[0]
		}
	}

	return &ShopUpdateRequest{
		Shop:     shop,
		NewLogo:  newLogo,
		NewCover: newCover,
	}, nil
}

// ToModel converts the request's shop to a domain model.
func (r *ShopUpdateRequest) ToModel() *models.Shop {
	return r.Shop.ToModel()
}

// Validate validates the shop update request.
func (r *ShopUpdateRequest) Validate() error {
	// Validate basic shop fields
	if err := r.validateBasicFields(); err != nil {
		return err
	}

	// Validate existing images have valid data
	if err := r.validateExistingImages(); err != nil {
		return err
	}

	// Validate new images (if any)
	if r.NewLogo != nil {
		if err := r.validateImage(r.NewLogo, "logo"); err != nil {
			return err
		}
	}
	if r.NewCover != nil {
		if err := r.validateImage(r.NewCover, "cover"); err != nil {
			return err
		}
	}

	return nil
}

func (r *ShopUpdateRequest) validateBasicFields() error {
	// Shop name is required
	if strings.TrimSpace(r.Shop.Name) == "" {
		return &httpErrors.BadRequestError{Message: "shop_name_is_required"}
	}

	// Shop slug is required
	if strings.TrimSpace(r.Shop.Slug) == "" {
		return &httpErrors.BadRequestError{Message: "shop_slug_is_required"}
	}

	return nil
}

func (r *ShopUpdateRequest) validateExistingImages() error {
	// Validate that existing images have valid IDs and URLs
	for _, img := range r.Shop.Images {
		if img.ID <= 0 {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_valid_id"}
		}
		if strings.TrimSpace(img.URL) == "" {
			return &httpErrors.BadRequestError{Message: "existing_image_must_have_url"}
		}
		// Validate type is logo or cover
		if img.Type != "logo" && img.Type != "cover" {
			return &httpErrors.BadRequestError{Message: "image_type_must_be_logo_or_cover"}
		}
	}
	return nil
}

func (r *ShopUpdateRequest) validateImage(imageHeader *multipart.FileHeader, imageType string) error {
	// Check file size (max 3MB per image)
	if imageHeader.Size > 3*1024*1024 {
		return &httpErrors.BadRequestError{Message: imageType + "_image_size_too_large_max_3mb"}
	}

	// Open file to check MIME type
	file, err := imageHeader.Open()
	if err != nil {
		return &httpErrors.BadRequestError{Message: "cannot_open_" + imageType + "_image_file"}
	}
	defer file.Close()

	// Read first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return &httpErrors.BadRequestError{Message: "cannot_read_" + imageType + "_image_file"}
	}

	// Check MIME type
	mimeType := http.DetectContentType(buffer)
	if !isValidImageType(mimeType) {
		return &httpErrors.BadRequestError{Message: "invalid_" + imageType + "_image_type_only_jpeg_png_allowed"}
	}

	// Reset file pointer for later use
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
		}
	}

	return nil
}

// ToLogoBuffer converts the logo FileHeader to byte slice for upload service.
// Returns nil if no new logo was provided.
func (r *ShopUpdateRequest) ToLogoBuffer() ([]byte, error) {
	if r.NewLogo == nil {
		return nil, nil
	}
	return r.fileHeaderToBuffer(r.NewLogo)
}

// ToCoverBuffer converts the cover FileHeader to byte slice for upload service.
// Returns nil if no new cover was provided.
func (r *ShopUpdateRequest) ToCoverBuffer() ([]byte, error) {
	if r.NewCover == nil {
		return nil, nil
	}
	return r.fileHeaderToBuffer(r.NewCover)
}

func (r *ShopUpdateRequest) fileHeaderToBuffer(imageHeader *multipart.FileHeader) ([]byte, error) {
	file, err := imageHeader.Open()
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}
	return buffer.Bytes(), nil
}
