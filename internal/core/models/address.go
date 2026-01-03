package models

// Address represents a physical location for a shop
type Address struct {
	ID      int     `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
}
