package models

type Option struct {
	ID       int     `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Quantity int     `json:"quantity,omitempty"`
	Order    int     `json:"order,omitempty"`
}
