package models

type SelectionType string

const (
	Single   SelectionType = "single"   // User can select only 1 option
	Multiple SelectionType = "multiple" // User can select multiple options (up to max_selections)
	Custom   SelectionType = "custom"   // Custom selection logic defined by user
)

type Variant struct {
	ID            int           `json:"id,omitempty"`
	Name          string        `json:"name,omitempty"`
	Order         int           `json:"order,omitempty"`
	SelectionType SelectionType `json:"selection_type,omitempty"`
	MaxSelections int           `json:"max_selections,omitempty"`
	IsRequired    bool          `json:"is_required,omitempty"`
	Options       []*Option     `json:"options,omitempty"`
}

// OptionsEqualForOrder compares selected options from frontend with database options
// Only compares options that were selected (present in frontendOptions)
func (v *Variant) OptionsEqualForOrder(frontendOptions []*Option) bool {
	for _, fo := range frontendOptions {
		// Find matching option in DB variant by ID
		dbOption := v.findOptionByID(fo.ID)
		if dbOption == nil {
			return false
		}
		// Compare option name and price
		if dbOption.Name != fo.Name || dbOption.Price != fo.Price {
			return false
		}
	}
	return true
}

// findOptionByID finds an option by ID in the variant's options
func (v *Variant) findOptionByID(id int) *Option {
	for _, o := range v.Options {
		if o.ID == id {
			return o
		}
	}
	return nil
}
