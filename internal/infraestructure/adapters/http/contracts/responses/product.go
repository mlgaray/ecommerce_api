package responses

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// ProductResponse represents a product in HTTP responses.
type ProductResponse struct {
	ID               int                       `json:"id,omitempty"`
	Name             string                    `json:"name,omitempty"`
	Description      string                    `json:"description,omitempty"`
	Price            float64                   `json:"price,omitempty"`
	Images           []*ProductImageResponse   `json:"images,omitempty"`
	Category         *ProductCategoryResponse  `json:"category,omitempty"`
	Variants         []*ProductVariantResponse `json:"variants"`
	IsActive         bool                      `json:"is_active"`
	IsPromotional    bool                      `json:"is_promotional"`
	PromotionalPrice float64                   `json:"promotional_price,omitempty"`
	IsHighlighted    bool                      `json:"is_highlighted"`
	IsStockeable     bool                      `json:"is_stockeable"`
	Stock            int                       `json:"stock"`
	MinimumStock     int                       `json:"minimum_stock,omitempty"`
	CreatedAt        time.Time                 `json:"created_at,omitempty"`
}

// ProductImageResponse represents an image in product responses.
type ProductImageResponse struct {
	ID  int    `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

// ProductCategoryResponse represents a category reference in product responses.
type ProductCategoryResponse struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProductVariantResponse represents a variant in product responses.
type ProductVariantResponse struct {
	ID            int                      `json:"id,omitempty"`
	Name          string                   `json:"name,omitempty"`
	Order         int                      `json:"order,omitempty"`
	SelectionType string                   `json:"selection_type,omitempty"`
	MaxSelections int                      `json:"max_selections,omitempty"`
	IsRequired    bool                     `json:"is_required,omitempty"`
	Options       []*ProductOptionResponse `json:"options,omitempty"`
}

// ProductOptionResponse represents an option in variant responses.
type ProductOptionResponse struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Price float64 `json:"price,omitempty"`
	Order int     `json:"order,omitempty"`
}

// ProductResponseFromModel converts a domain Product to a ProductResponse.
func ProductResponseFromModel(p *models.Product) *ProductResponse {
	if p == nil {
		return nil
	}

	response := &ProductResponse{
		ID:               p.ID,
		Name:             p.Name,
		Description:      p.Description,
		Price:            p.Price,
		IsActive:         p.IsActive,
		IsPromotional:    p.IsPromotional,
		PromotionalPrice: p.PromotionalPrice,
		IsHighlighted:    p.IsHighlighted,
		IsStockeable:     p.IsStockeable,
		Stock:            p.Stock,
		MinimumStock:     p.MinimumStock,
		CreatedAt:        p.CreatedAt,
	}

	// Convert images
	if p.Images != nil {
		response.Images = make([]*ProductImageResponse, len(p.Images))
		for i, img := range p.Images {
			response.Images[i] = &ProductImageResponse{
				ID:  img.ID,
				URL: img.URL,
			}
		}
	}

	// Convert category
	if p.Category != nil {
		response.Category = &ProductCategoryResponse{
			ID:          p.Category.ID,
			Name:        p.Category.Name,
			Description: p.Category.Description,
		}
	}

	// Convert variants
	if p.Variants != nil {
		response.Variants = make([]*ProductVariantResponse, len(p.Variants))
		for i, v := range p.Variants {
			response.Variants[i] = variantResponseFromModel(v)
		}
	}

	return response
}

// variantResponseFromModel converts a domain Variant to a ProductVariantResponse.
func variantResponseFromModel(v *models.Variant) *ProductVariantResponse {
	if v == nil {
		return nil
	}

	response := &ProductVariantResponse{
		ID:            v.ID,
		Name:          v.Name,
		Order:         v.Order,
		SelectionType: string(v.SelectionType),
		MaxSelections: v.MaxSelections,
		IsRequired:    v.IsRequired,
	}

	// Convert options
	if v.Options != nil {
		response.Options = make([]*ProductOptionResponse, len(v.Options))
		for i, opt := range v.Options {
			response.Options[i] = &ProductOptionResponse{
				ID:    opt.ID,
				Name:  opt.Name,
				Price: opt.Price,
				Order: opt.Order,
			}
		}
	}

	return response
}

// ProductResponsesFromModels converts a slice of domain Products to ProductResponses.
func ProductResponsesFromModels(products []*models.Product) []*ProductResponse {
	if products == nil {
		return nil
	}

	responses := make([]*ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ProductResponseFromModel(p)
	}
	return responses
}

// CreateProductResponse represents the response for product creation.
type CreateProductResponse struct {
	Product *ProductResponse `json:"product"`
}

// GetProductResponse represents the response for product retrieval.
type GetProductResponse struct {
	Product *ProductResponse `json:"product"`
}

// UpdateProductResponse represents the response for a product update.
type UpdateProductResponse struct {
	Product *ProductResponse `json:"product"`
}

// ListProductsResponse represents the HTTP response for list products.
type ListProductsResponse struct {
	Products   []*ProductResponse `json:"products"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
	TotalCount *int               `json:"total_count,omitempty"`
}

// NewListProductsResponse creates a ListProductsResponse from domain models.
func NewListProductsResponse(
	products []*models.Product,
	nextCursor string,
	hasMore bool,
	totalCount *int,
) *ListProductsResponse {
	return &ListProductsResponse{
		Products:   ProductResponsesFromModels(products),
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}
}
