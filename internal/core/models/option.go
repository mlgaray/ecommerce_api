package models

type Option struct {
	ID    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Price float64 `json:"price,omitempty"`
	Order int     `json:"order,omitempty"`
}
