package models

import "github.com/mlgaray/ecommerce_api/internal/core/errors"

type Product struct {
	ID               int            `json:"id,omitempty"`
	Name             string         `json:"name,omitempty"`
	Description      string         `json:"description,omitempty"`
	Price            float64        `json:"price,omitempty"`
	Images           []ProductImage `json:"images,omitempty"`
	Category         *Category      `json:"category,omitempty"`
	Variants         []*Variant     `json:"variants"`
	IsActive         bool           `json:"is_active"`
	IsPromotional    bool           `json:"is_promotional"`
	PromotionalPrice float64        `json:"promotional_price,omitempty"`
	IsHighlighted    bool           `json:"is_highlighted"`
	Stock            int            `json:"stock"`
	MinimumStock     int            `json:"minimum_stock,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (p *Product) GetID() int {
	return p.ID
}

// Validate validates business rules for the Product domain model
func (p *Product) Validate() error {
	if err := p.validatePriceAndStock(); err != nil {
		return err
	}

	if err := p.validateMinimumStock(); err != nil {
		return err
	}

	if err := p.validatePromotionalPrice(); err != nil {
		return err
	}

	return nil
}

// validatePriceAndStock validates basic price and stock business rules
func (p *Product) validatePriceAndStock() error {
	// Business rule: price must be positive
	if p.Price <= 0 {
		return &errors.ValidationError{
			Message: errors.ProductPriceMustBePositive,
		}
	}

	// Business rule: stock cannot be negative
	if p.Stock < 0 {
		return &errors.ValidationError{
			Message: errors.ProductStockCannotBeNegative,
		}
	}

	return nil
}

// validateMinimumStock validates minimum stock business rules
func (p *Product) validateMinimumStock() error {
	// Business rule: minimum stock cannot be negative
	if p.MinimumStock < 0 {
		return &errors.ValidationError{
			Message: errors.ProductMinimumStockCannotBeNegative,
		}
	}

	// Business rule: minimum stock can only exist if there's stock
	if p.MinimumStock > 0 && p.Stock == 0 {
		return &errors.ValidationError{
			Message: errors.MinimumStockRequiresStock,
		}
	}

	// Business rule: minimum stock cannot be greater than stock
	if p.Stock > 0 && p.MinimumStock > p.Stock {
		return &errors.ValidationError{
			Message: errors.ProductMinimumStockCannotBeGreaterThanStock,
		}
	}

	return nil
}

// validatePromotionalPrice validates promotional price business rules
func (p *Product) validatePromotionalPrice() error {
	// Business rule: if promotional, must have promotional price
	if p.IsPromotional && p.PromotionalPrice <= 0 {
		return &errors.ValidationError{
			Message: errors.PromotionalProductRequiresPromotionalPrice,
		}
	}

	// Business rule: promotional price must be lower than regular price
	if p.IsPromotional && p.PromotionalPrice >= p.Price {
		return &errors.ValidationError{
			Message: errors.PromotionalPriceMustBeLowerThanRegularPrice,
		}
	}

	return nil
}

// CanBeSold checks if the product can be sold (business logic)
func (p *Product) CanBeSold() bool {
	return p.IsActive && p.Stock > 0
}

// IsLowStock checks if stock is below minimum threshold (business logic)
func (p *Product) IsLowStock() bool {
	return p.Stock <= p.MinimumStock
}

// GetEffectivePrice returns the current price considering promotions (business logic)
func (p *Product) GetEffectivePrice() float64 {
	if p.IsPromotional {
		return p.PromotionalPrice
	}
	return p.Price
}

// DecrementStock reduces stock by given quantity (business logic with validation)
func (p *Product) DecrementStock(quantity int) error {
	if quantity <= 0 {
		return &errors.ValidationError{
			Message: "quantity_must_be_positive",
		}
	}

	if p.Stock < quantity {
		return &errors.BusinessRuleError{
			Message: "insufficient_stock",
		}
	}

	p.Stock -= quantity
	return nil
}

// IncrementStock increases stock by given quantity (business logic with validation)
func (p *Product) IncrementStock(quantity int) error {
	if quantity <= 0 {
		return &errors.ValidationError{
			Message: "quantity_must_be_positive",
		}
	}

	p.Stock += quantity
	return nil
}
