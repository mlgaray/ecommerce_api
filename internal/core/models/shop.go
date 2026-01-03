package models

type Shop struct {
	ID                 int                  `json:"id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Slug               string               `json:"slug,omitempty"`
	Email              string               `json:"email,omitempty"`
	Phone              string               `json:"phone,omitempty"`
	Instagram          string               `json:"instagram,omitempty"`
	Images             []*Image             `json:"images,omitempty"` // type='logo' or type='cover'
	Address            *Address             `json:"address,omitempty"`
	PaymentMethods     []*PaymentMethod     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*DeliveryMethod    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*OperatingSchedule `json:"operating_schedules,omitempty"`
}
