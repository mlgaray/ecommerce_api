package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

type Product struct {
	ID               int        `json:"id,omitempty"`
	Name             string     `json:"name,omitempty"`
	Description      string     `json:"description,omitempty"`
	Price            float64    `json:"price,omitempty"`
	Images           []*Image   `json:"images,omitempty"`
	Category         *Category  `json:"category,omitempty"`
	Variants         []*Variant `json:"variants"`
	IsActive         bool       `json:"is_active"`
	IsPromotional    bool       `json:"is_promotional"`
	PromotionalPrice float64    `json:"promotional_price,omitempty"`
	IsHighlighted    bool       `json:"is_highlighted"`
	IsStockeable     bool       `json:"is_stockeable"`
	Stock            int        `json:"stock"`
	MinimumStock     int        `json:"minimum_stock,omitempty"`
	CreatedAt        time.Time  `json:"created_at,omitempty"` // Timestamp de creación
}

// GetID implements Identifiable interface for pagination
func (p *Product) GetID() int {
	return p.ID
}

// GetSortValue implements Sortable interface for pagination
// Returns the value of the field being sorted by, or nil if not supported
func (p *Product) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case SortByPrice:
		return p.Price
	case SortByName:
		return p.Name
	case SortByCreatedAt:
		return p.CreatedAt
	case SortByID:
		return nil // When sorting by ID, no additional sort value needed
	default:
		return nil // Unsupported sort field
	}
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
	// Skip validation if product doesn't manage stock
	if !p.IsStockeable {
		return nil
	}

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
// Returns true if product is active and either doesn't manage stock or has stock available
func (p *Product) CanBeSold() bool {
	if !p.IsActive {
		return false
	}
	// If product doesn't manage stock, it can always be sold
	if !p.IsStockeable {
		return true
	}
	return p.Stock > 0
}

// IsLowStock checks if stock is below minimum threshold (business logic)
// Returns false if product doesn't manage stock
func (p *Product) IsLowStock() bool {
	if !p.IsStockeable {
		return false
	}
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
// If product doesn't manage stock, this is a no-op
func (p *Product) DecrementStock(quantity int) error {
	// Products that don't manage stock don't need decrement
	if !p.IsStockeable {
		return nil
	}

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
// If product doesn't manage stock, this is a no-op
func (p *Product) IncrementStock(quantity int) error {
	// Products that don't manage stock don't need increment
	if !p.IsStockeable {
		return nil
	}

	if quantity <= 0 {
		return &errors.ValidationError{
			Message: "quantity_must_be_positive",
		}
	}

	p.Stock += quantity
	return nil
}

// HasSufficientStock checks if product has enough stock for the given quantity
// Returns true if product doesn't manage stock or has sufficient stock
func (p *Product) HasSufficientStock(quantity int) bool {
	if !p.IsStockeable {
		return true
	}
	return p.Stock >= quantity
}

// EqualsForOrder compares product data sent from frontend with database product
// Used to validate that frontend data matches current backend data
// Compares: Name, Price, IsPromotional, PromotionalPrice
func (p *Product) EqualsForOrder(other *Product) bool {
	if other == nil {
		return false
	}
	return p.Name == other.Name &&
		p.Price == other.Price &&
		p.IsPromotional == other.IsPromotional &&
		p.PromotionalPrice == other.PromotionalPrice
}

// VariantsEqualForOrder compares selected variants from frontend with database variants
// Only compares variants that were selected (present in frontendVariants)
func (p *Product) VariantsEqualForOrder(frontendVariants []*Variant) bool {
	for _, fv := range frontendVariants {
		// Find matching variant in DB product by ID
		dbVariant := p.findVariantByID(fv.ID)
		if dbVariant == nil {
			return false
		}
		// Compare variant name
		if dbVariant.Name != fv.Name {
			return false
		}
		// Compare selected options
		if !dbVariant.OptionsEqualForOrder(fv.Options) {
			return false
		}
	}
	return true
}

// findVariantByID finds a variant by ID in the product's variants
func (p *Product) findVariantByID(id int) *Variant {
	for _, v := range p.Variants {
		if v.ID == id {
			return v
		}
	}
	return nil
}
