package responses

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CategoryResponse represents a category in HTTP responses.
type CategoryResponse struct {
	ID          int                    `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Image       *CategoryImageResponse `json:"image,omitempty"`
	CreatedAt   time.Time              `json:"created_at,omitempty"`
}

// CategoryImageResponse represents an image in category responses.
type CategoryImageResponse struct {
	ID  int    `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

// CategoryResponseFromModel converts a domain Category to a CategoryResponse.
func CategoryResponseFromModel(c *models.Category) *CategoryResponse {
	if c == nil {
		return nil
	}

	response := &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}

	if c.Image != nil {
		response.Image = &CategoryImageResponse{
			ID:  c.Image.ID,
			URL: c.Image.URL,
		}
	}

	return response
}

// CategoryResponsesFromModels converts a slice of domain Categories to CategoryResponses.
func CategoryResponsesFromModels(categories []*models.Category) []*CategoryResponse {
	if categories == nil {
		return nil
	}

	responses := make([]*CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = CategoryResponseFromModel(c)
	}
	return responses
}

// CreateCategoryResponse represents the response for category creation.
type CreateCategoryResponse struct {
	Category *CategoryResponse `json:"category"`
}

// GetCategoryResponse represents the response for category retrieval.
type GetCategoryResponse struct {
	Category *CategoryResponse `json:"category"`
}

// UpdateCategoryResponse represents the response for a category update.
type UpdateCategoryResponse struct {
	Category *CategoryResponse `json:"category"`
}

// ListCategoriesResponse represents the HTTP response for list categories.
type ListCategoriesResponse struct {
	Categories []*CategoryResponse `json:"categories"`
	NextCursor string              `json:"next_cursor,omitempty"`
	HasMore    bool                `json:"has_more"`
	TotalCount *int                `json:"total_count,omitempty"`
}

// NewListCategoriesResponse creates a ListCategoriesResponse from domain models.
func NewListCategoriesResponse(
	categories []*models.Category,
	nextCursor string,
	hasMore bool,
	totalCount *int,
) *ListCategoriesResponse {
	return &ListCategoriesResponse{
		Categories: CategoryResponsesFromModels(categories),
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}
}
