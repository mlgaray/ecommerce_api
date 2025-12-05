package contracts

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// PaginatedCategoriesResponse represents the HTTP response for paginated categories.
type PaginatedCategoriesResponse struct {
	Categories []*models.Category `json:"categories"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
	TotalCount *int               `json:"total_count,omitempty"`
}
