package models

type Category struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
}
