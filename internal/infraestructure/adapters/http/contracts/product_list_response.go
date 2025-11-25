package contracts

import "github.com/mlgaray/ecommerce_api/internal/core/models"

// PaginatedProductsResponse represents the HTTP response for paginated products
type PaginatedProductsResponse struct {
	Products   []*models.Product `json:"products"`
	NextCursor string            `json:"next_cursor,omitempty"` // Opaque base64-encoded cursor
	HasMore    bool              `json:"has_more"`
	TotalCount *int              `json:"total_count,omitempty"` // Total count (only on first page when cursor is empty, nil on subsequent pages)
}
