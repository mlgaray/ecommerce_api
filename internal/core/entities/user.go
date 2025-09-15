package entities

type User struct {
	ID       int     `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	Email    string  `json:"email,omitempty"`
	Password string  `json:"password,omitempty"`
	Roles    []*Role `json:"roles,omitempty"`
}
