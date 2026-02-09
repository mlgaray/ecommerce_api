package models

// Customer representa los datos del cliente de una orden
type Customer struct {
	Name    string   `json:"name,omitempty"`
	Phone   string   `json:"phone,omitempty"`
	Email   string   `json:"email,omitempty"`
	Address *Address `json:"address,omitempty"`
}
