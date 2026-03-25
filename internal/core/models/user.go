package models

type User struct {
	ID           int     `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	LastName     string  `json:"last_name,omitempty"`
	Email        string  `json:"email,omitempty"`
	Password     string  `json:"password,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	IsActive     bool    `json:"is_active,omitempty"`
	ProfileImage *Image  `json:"profile_image,omitempty"`
	Roles        []*Role `json:"roles,omitempty"`
}
