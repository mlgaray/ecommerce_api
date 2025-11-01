package models

type Product struct {
	ID               int        `json:"id,omitempty"`
	Name             string     `json:"name,omitempty"`
	Description      string     `json:"description,omitempty"`
	Price            float64    `json:"price,omitempty"`
	Images           []string   `json:"images,omitempty"`
	Category         *Category  `json:"category,omitempty"`
	Variants         []*Variant `json:"variants"`
	IsActive         bool       `json:"is_active"`
	IsPromotional    bool       `json:"is_promotional"`
	PromotionalPrice float64    `json:"promotional_price,omitempty"`
	IsHighlighted    bool       `json:"is_highlighted"`
	Stock            int        `json:"stock"`
	MinimumStock     int        `json:"minimum_stock,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (p *Product) GetID() int {
	return p.ID
}
