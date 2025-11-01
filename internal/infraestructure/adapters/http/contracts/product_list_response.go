package contracts

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// PaginatedProductsResponse represents the HTTP response for paginated products
type PaginatedProductsResponse struct {
	Products   []*models.Product `json:"products"`
	NextCursor int               `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}
