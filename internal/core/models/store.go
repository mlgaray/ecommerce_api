package models

import "time"

// Store represents the public-facing view of a shop for customers.
// Similar to Shop but can have additional calculated fields in the future (e.g., IsOpen).
type Store struct {
	ID                 int                  `json:"id,omitempty"`
	Name               string               `json:"name,omitempty"`
	Slug               string               `json:"slug,omitempty"`
	Email              string               `json:"email,omitempty"`
	Phone              string               `json:"phone,omitempty"`
	Instagram          string               `json:"instagram,omitempty"`
	Images             []*Image             `json:"images,omitempty"`
	Address            *Address             `json:"address,omitempty"`
	PaymentMethods     []*PaymentMethod     `json:"payment_methods,omitempty"`
	DeliveryMethods    []*DeliveryMethod    `json:"delivery_methods,omitempty"`
	OperatingSchedules []*OperatingSchedule `json:"operating_schedules,omitempty"`
	Timezone           *Timezone            `json:"timezone,omitempty"`
	IsOpen             bool                 `json:"is_open"`
}

// NewStoreFromShop creates a Store from a Shop model.
func NewStoreFromShop(shop *Shop) *Store {
	if shop == nil {
		return nil
	}

	store := &Store{
		ID:                 shop.ID,
		Name:               shop.Name,
		Slug:               shop.Slug,
		Email:              shop.Email,
		Phone:              shop.Phone,
		Instagram:          shop.Instagram,
		Images:             shop.Images,
		Address:            shop.Address,
		PaymentMethods:     shop.PaymentMethods,
		DeliveryMethods:    shop.DeliveryMethods,
		OperatingSchedules: shop.OperatingSchedules,
		Timezone:           shop.Timezone,
	}

	store.IsOpen = store.CalculateIsOpen(time.Now())

	return store
}

// CalculateIsOpen determines if the store is currently open based on
// the store's timezone and operating schedules.
// Returns false if no timezone is configured or no matching schedule is found.
func (s *Store) CalculateIsOpen(now time.Time) bool {
	if s.Timezone == nil || s.Timezone.Identifier == "" {
		return false
	}

	if len(s.OperatingSchedules) == 0 {
		return false
	}

	// Load the store's timezone
	loc, err := time.LoadLocation(s.Timezone.Identifier)
	if err != nil {
		return false
	}

	// Convert current time to store's local time
	localTime := now.In(loc)
	dayOfWeek := DayOfWeekFromTime(localTime.Weekday())

	// Check if any schedule matches the current day and time
	for _, schedule := range s.OperatingSchedules {
		if schedule.DayOfWeek == dayOfWeek && schedule.IsTimeWithinRange(localTime) {
			return true
		}
	}

	return false
}
