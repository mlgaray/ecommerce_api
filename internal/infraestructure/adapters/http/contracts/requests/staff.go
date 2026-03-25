package requests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/pagination"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Create Staff
// =============================================================================

type CreateStaffRequest struct {
	Name         string                `json:"name"`
	LastName     string                `json:"last_name"`
	Email        string                `json:"email"`
	Password     string                `json:"password"`
	RoleID       int                   `json:"role_id"`
	ProfileImage *multipart.FileHeader `json:"-"`
}

func NewCreateStaffRequest(r *http.Request) (*CreateStaffRequest, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	roleIDStr := r.FormValue("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		return nil, &httpErrors.BadRequestError{Message: "invalid_role_id"}
	}

	var imageHeader *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["profile_image"]; exists && len(files) > 0 {
			imageHeader = files[0]
		}
	}

	return &CreateStaffRequest{
		Name:         name,
		LastName:     lastName,
		Email:        email,
		Password:     password,
		RoleID:       roleID,
		ProfileImage: imageHeader,
	}, nil
}

func (r *CreateStaffRequest) Validate() error {
	if r.Name == "" {
		return &httpErrors.BadRequestError{Message: "staff_name_is_required"}
	}
	if r.LastName == "" {
		return &httpErrors.BadRequestError{Message: "staff_last_name_is_required"}
	}
	if r.Email == "" {
		return &httpErrors.BadRequestError{Message: "staff_email_is_required"}
	}
	if !emailRegex.MatchString(r.Email) {
		return &httpErrors.BadRequestError{Message: "invalid_email_format"}
	}
	if strings.TrimSpace(r.Password) == "" {
		return &httpErrors.BadRequestError{Message: "staff_password_is_required"}
	}
	if len(r.Password) < 6 {
		return &httpErrors.BadRequestError{Message: "password_must_be_at_least_6_characters"}
	}
	if r.ProfileImage != nil {
		if err := validateStaffImage(r.ProfileImage); err != nil {
			return err
		}
	}
	return nil
}

func (r *CreateStaffRequest) ToUserModel() *models.User {
	return &models.User{
		Name:     r.Name,
		LastName: r.LastName,
		Email:    r.Email,
		Password: r.Password,
	}
}

func (r *CreateStaffRequest) ToImageBuffer() ([]byte, error) {
	if r.ProfileImage == nil {
		return nil, nil
	}
	return staffFileHeaderToBuffer(r.ProfileImage)
}

// =============================================================================
// Update Staff
// =============================================================================

type UpdateStaffRequest struct {
	Name         string                `json:"name"`
	LastName     string                `json:"last_name"`
	Email        string                `json:"email"`
	Password     string                `json:"password"`
	RoleID       int                   `json:"role_id"`
	ProfileImage *multipart.FileHeader `json:"-"`
}

func NewUpdateStaffRequest(r *http.Request) (*UpdateStaffRequest, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	roleIDStr := r.FormValue("role_id")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		return nil, &httpErrors.BadRequestError{Message: "invalid_role_id"}
	}

	var imageHeader *multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if files, exists := r.MultipartForm.File["profile_image"]; exists && len(files) > 0 {
			imageHeader = files[0]
		}
	}

	return &UpdateStaffRequest{
		Name:         name,
		LastName:     lastName,
		Email:        email,
		Password:     password,
		RoleID:       roleID,
		ProfileImage: imageHeader,
	}, nil
}

func (r *UpdateStaffRequest) Validate() error {
	if r.Name == "" {
		return &httpErrors.BadRequestError{Message: "staff_name_is_required"}
	}
	if r.LastName == "" {
		return &httpErrors.BadRequestError{Message: "staff_last_name_is_required"}
	}
	if r.Email == "" {
		return &httpErrors.BadRequestError{Message: "staff_email_is_required"}
	}
	if !emailRegex.MatchString(r.Email) {
		return &httpErrors.BadRequestError{Message: "invalid_email_format"}
	}
	if r.Password != "" && len(r.Password) < 6 {
		return &httpErrors.BadRequestError{Message: "password_must_be_at_least_6_characters"}
	}
	if r.ProfileImage != nil {
		if err := validateStaffImage(r.ProfileImage); err != nil {
			return err
		}
	}
	return nil
}

func (r *UpdateStaffRequest) ToUserModel() *models.User {
	return &models.User{
		Name:     r.Name,
		LastName: r.LastName,
		Email:    r.Email,
		Password: r.Password,
	}
}

func (r *UpdateStaffRequest) ToImageBuffer() ([]byte, error) {
	if r.ProfileImage == nil {
		return nil, nil
	}
	return staffFileHeaderToBuffer(r.ProfileImage)
}

// =============================================================================
// Shared helpers
// =============================================================================

func isValidStaffImageType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}
	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}

func validateStaffImage(image *multipart.FileHeader) error {
	if image.Size > 3*1024*1024 {
		return &httpErrors.BadRequestError{Message: "image_size_too_large_max_3mb"}
	}

	file, err := image.Open()
	if err != nil {
		return &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	mimeType := http.DetectContentType(buffer)
	if !isValidStaffImageType(mimeType) {
		return &httpErrors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
	}

	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return &httpErrors.BadRequestError{Message: "cannot_reset_file_pointer"}
		}
	}

	return nil
}

func staffFileHeaderToBuffer(fh *multipart.FileHeader) ([]byte, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_open_image_file"}
	}
	defer file.Close()

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		return nil, &httpErrors.BadRequestError{Message: "cannot_read_image_file"}
	}

	return buf.Bytes(), nil
}

// =============================================================================
// Filters (pagination + search)
// =============================================================================

// StaffFiltersRequest represents HTTP query parameters for staff listing.
type StaffFiltersRequest struct {
	Search    *string
	IsActive  *string
	SortBy    string
	SortOrder string
	Limit     int
	Cursor    string
}

// NewStaffFiltersRequest parses query parameters from URL.
func NewStaffFiltersRequest(queryParams map[string][]string) (*StaffFiltersRequest, error) {
	request := &StaffFiltersRequest{}

	if search := getQueryParam(queryParams, "search"); search != "" {
		request.Search = &search
	}

	if isActive := getQueryParam(queryParams, "is_active"); isActive != "" {
		request.IsActive = &isActive
	}

	request.SortBy = getQueryParam(queryParams, "sort")
	request.SortOrder = getQueryParam(queryParams, "order")

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		request.Limit = limit
	}

	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		request.Cursor = cursorStr
	}

	return request, nil
}

// Validate validates HTTP-specific concerns.
func (r *StaffFiltersRequest) Validate() error {
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		*r.Search = trimmed
	}

	if r.IsActive != nil {
		if _, err := strconv.ParseBool(*r.IsActive); err != nil {
			return &httpErrors.BadRequestError{Message: "is_active_must_be_true_or_false"}
		}
	}

	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	return nil
}

// ToStaffFilters converts HTTP request to domain model.
func (r *StaffFiltersRequest) ToStaffFilters() models.StaffFilters {
	filters := models.StaffFilters{
		Search:    r.Search,
		SortBy:    r.SortBy,
		SortOrder: r.SortOrder,
		Limit:     r.Limit,
	}

	if r.IsActive != nil {
		isActive, _ := strconv.ParseBool(*r.IsActive)
		filters.IsActive = &isActive
	}

	if r.Cursor != "" {
		cursorData, err := pagination.DecodeCursor(r.Cursor)
		if err != nil {
			return filters
		}
		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}
