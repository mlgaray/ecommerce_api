package models

import (
	"regexp"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// hexColorRegex validates hex color format: # followed by exactly 6 hex digits
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type Shop struct {
	ID                 int                  `json:"id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Slug               string               `json:"slug,omitempty"`
	Email              string               `json:"email,omitempty"`
	Phone              string               `json:"phone,omitempty"`
	Instagram          string               `json:"instagram,omitempty"`
	PrimaryColor       *string              `json:"primary_color,omitempty"`
	Images             []*Image             `json:"images,omitempty"` // type='logo' or type='cover'
	Address            *Address             `json:"address,omitempty"`
	PaymentMethods     []*PaymentMethod     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*DeliveryMethod    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*OperatingSchedule `json:"operating_schedules,omitempty"`
	Timezone           *Timezone            `json:"timezone,omitempty"`
}

// Validate validates the Shop domain model.
// Returns a ValidationError if primary_color is set but has invalid hex format.
func (s *Shop) Validate() error {
	if s.PrimaryColor != nil && *s.PrimaryColor != "" {
		if !hexColorRegex.MatchString(*s.PrimaryColor) {
			return &errors.ValidationError{Message: errors.InvalidPrimaryColorFormat}
		}
	}
	return nil
}
