package models

type (
	Shop struct {
		ID     int    `json:"id,omitempty"`
		UserID int    `json:"user_id,omitempty"`
		Name   string `json:"name,omitempty"`
		Slug   string `json:"slug,omitempty"`
		Email  string `json:"email,omitempty"`
		Phone  string `json:"phone,omitempty"`
		// Categories []*Category `json:"categories,omitempty"`
		Instagram string `json:"instagram,omitempty"`
		// Address    *Address    `json:"address,omitempty"`
		Image string `json:"image,omitempty"`
		File  []byte `json:"file,omitempty"`
	}
)
