package models

type Category struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Image       *Image `json:"image,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (c *Category) GetID() int {
	return c.ID
}
