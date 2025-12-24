package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// DayOfWeek represents the day of the week (0=Sunday, 6=Saturday)
// Compatible with Go's time.Weekday
type DayOfWeek int

const (
	Sunday    DayOfWeek = 0
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
)

// OperatingSchedule represents a time range for a specific day of the week
// A shop can have multiple schedules per day (split hours)
type OperatingSchedule struct {
	ID        int       `json:"id,omitempty"`
	ShopID    int       `json:"shop_id,omitempty"`
	DayOfWeek DayOfWeek `json:"day_of_week"`
	OpenTime  string    `json:"open_time"`  // Format: "HH:MM"
	CloseTime string    `json:"close_time"` // Format: "HH:MM"
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Validate validates the operating schedule business rules
func (os *OperatingSchedule) Validate() error {
	if os.DayOfWeek < 0 || os.DayOfWeek > 6 {
		return &errors.ValidationError{Message: errors.InvalidDayOfWeek}
	}
	if os.OpenTime == "" {
		return &errors.ValidationError{Message: errors.OpenTimeRequired}
	}
	if os.CloseTime == "" {
		return &errors.ValidationError{Message: errors.CloseTimeRequired}
	}

	// Parse times and validate close > open
	openTime, err := time.Parse("15:04", os.OpenTime)
	if err != nil {
		return &errors.ValidationError{Message: errors.InvalidTimeFormat}
	}
	closeTime, err := time.Parse("15:04", os.CloseTime)
	if err != nil {
		return &errors.ValidationError{Message: errors.InvalidTimeFormat}
	}

	if !closeTime.After(openTime) {
		return &errors.ValidationError{Message: errors.InvalidTimeRange}
	}

	return nil
}

// IsTimeWithinRange checks if a given time falls within this schedule's range
// Only compares hours and minutes, ignores the date
func (os *OperatingSchedule) IsTimeWithinRange(t time.Time) bool {
	openTime, err := time.Parse("15:04", os.OpenTime)
	if err != nil {
		return false
	}
	closeTime, err := time.Parse("15:04", os.CloseTime)
	if err != nil {
		return false
	}

	// Extract only hour and minute from input time
	checkTime, _ := time.Parse("15:04", t.Format("15:04"))

	return !checkTime.Before(openTime) && !checkTime.After(closeTime)
}

// DayOfWeekFromTime converts a time.Weekday to DayOfWeek
func DayOfWeekFromTime(w time.Weekday) DayOfWeek {
	return DayOfWeek(w)
}
