package models

type SelectionType string

const (
	Single    SelectionType = "single"
	Unlimited SelectionType = "unlimited"
	Custom    SelectionType = "custom"
)

type Variant struct {
	ID            int           `json:"id,omitempty"`
	Name          string        `json:"name,omitempty"`
	Order         int           `json:"order,omitempty"`
	SelectionType SelectionType `json:"selection_type,omitempty"`
	MaxSelections int           `json:"max_selections,omitempty"`
	Options       []*Option     `json:"options,omitempty"`
}
