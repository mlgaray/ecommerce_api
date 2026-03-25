package responses

import (
	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// StaffResponse represents a staff member in HTTP responses.
type StaffResponse struct {
	ID       int                  `json:"id"`
	IsActive bool                 `json:"is_active"`
	User     *StaffUserResponse   `json:"user"`
	Roles    []*StaffRoleResponse `json:"roles,omitempty"`
}

// StaffUserResponse represents the user data within a staff response.
type StaffUserResponse struct {
	ID           int                        `json:"id"`
	Name         string                     `json:"name"`
	LastName     string                     `json:"last_name"`
	Email        string                     `json:"email"`
	ProfileImage *StaffProfileImageResponse `json:"profile_image,omitempty"`
}

// StaffProfileImageResponse represents a profile image in staff responses.
type StaffProfileImageResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// StaffRoleResponse represents a role in staff responses.
type StaffRoleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// StaffResponseFromModel converts a domain Staff to a StaffResponse.
func StaffResponseFromModel(m *models.Staff) *StaffResponse {
	if m == nil {
		return nil
	}

	response := &StaffResponse{
		ID:       m.ID,
		IsActive: m.IsActive,
	}

	if m.User != nil {
		userResp := &StaffUserResponse{
			ID:       m.User.ID,
			Name:     m.User.Name,
			LastName: m.User.LastName,
			Email:    m.User.Email,
		}
		if m.User.ProfileImage != nil {
			userResp.ProfileImage = &StaffProfileImageResponse{
				ID:  m.User.ProfileImage.ID,
				URL: m.User.ProfileImage.URL,
			}
		}
		response.User = userResp
	}

	if m.Roles != nil {
		roles := make([]*StaffRoleResponse, len(m.Roles))
		for i, role := range m.Roles {
			roles[i] = &StaffRoleResponse{
				ID:   role.ID,
				Name: role.Name,
			}
		}
		response.Roles = roles
	}

	return response
}

// StaffResponsesFromModels converts a slice of domain Staff to StaffResponses.
func StaffResponsesFromModels(staff []*models.Staff) []*StaffResponse {
	if staff == nil {
		return nil
	}

	responses := make([]*StaffResponse, len(staff))
	for i, m := range staff {
		responses[i] = StaffResponseFromModel(m)
	}
	return responses
}

// CreateStaffResponse represents the response for staff creation.
type CreateStaffResponse struct {
	Staff *StaffResponse `json:"staff"`
}

// GetStaffResponse represents the response for staff retrieval.
type GetStaffResponse struct {
	Staff *StaffResponse `json:"staff"`
}

// UpdateStaffResponse represents the response for staff update.
type UpdateStaffResponse struct {
	Staff *StaffResponse `json:"staff"`
}

// ListStaffResponse represents the HTTP response for list staff with pagination.
type ListStaffResponse struct {
	Staff      []*StaffResponse `json:"staff"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
	TotalCount *int             `json:"total_count,omitempty"`
}

// NewListStaffResponse creates a ListStaffResponse from domain models.
func NewListStaffResponse(
	staff []*models.Staff,
	nextCursor string,
	hasMore bool,
	totalCount *int,
) *ListStaffResponse {
	return &ListStaffResponse{
		Staff:      StaffResponsesFromModels(staff),
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}
}

// ToggleStaffStatusResponse represents the response for toggling staff status.
type ToggleStaffStatusResponse struct {
	Staff *StaffResponse `json:"staff"`
}
