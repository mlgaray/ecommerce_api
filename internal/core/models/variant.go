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
	Options       []*Option     `json:"options,omitempty"`
}
