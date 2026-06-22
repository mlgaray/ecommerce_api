package responses

import "github.com/mlgaray/ecommerce_api/internal/core/models"

type RoleResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListRolesResponse struct {
	Roles []*RoleResponse `json:"roles"`
}

func NewListRolesResponse(roles []*models.Role) *ListRolesResponse {
	resp := make([]*RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = &RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
		}
	}
	return &ListRolesResponse{Roles: resp}
}
