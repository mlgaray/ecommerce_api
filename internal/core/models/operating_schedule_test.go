package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func newValidOperatingSchedule() *OperatingSchedule {
	return &OperatingSchedule{
		DayOfWeek: Monday,
		OpenTime:  "09:00",
		CloseTime: "18:00",
	}
}

// =============================================================================
// Validate Tests
// =============================================================================

func TestOperatingSchedule_Validate(t *testing.T) {
	t.Run("when schedule is valid then returns nil", func(t *testing.T) {
		schedule := newValidOperatingSchedule()

		err := schedule.Validate()

		assert.NoError(t, err)
	})

	t.Run("when day_of_week is negative then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.DayOfWeek = -1

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidDayOfWeek, validationErr.Message)
	})

	t.Run("when day_of_week is greater than 6 then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.DayOfWeek = 7

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidDayOfWeek, validationErr.Message)
	})

	t.Run("when open_time is empty then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.OpenTime = ""

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.OpenTimeRequired, validationErr.Message)
	})

	t.Run("when close_time is empty then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.CloseTime = ""

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CloseTimeRequired, validationErr.Message)
	})

	t.Run("when open_time has invalid format then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.OpenTime = "invalid"

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidTimeFormat, validationErr.Message)
	})

	t.Run("when close_time has invalid format then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.CloseTime = "25:00"

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidTimeFormat, validationErr.Message)
	})

	t.Run("when close_time is before open_time then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.OpenTime = "18:00"
		schedule.CloseTime = "09:00"

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidTimeRange, validationErr.Message)
	})

	t.Run("when close_time equals open_time then returns error", func(t *testing.T) {
		schedule := newValidOperatingSchedule()
		schedule.OpenTime = "09:00"
		schedule.CloseTime = "09:00"

		err := schedule.Validate()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidTimeRange, validationErr.Message)
	})

	t.Run("when all days of week are valid then returns nil", func(t *testing.T) {
		days := []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}

		for _, day := range days {
			schedule := newValidOperatingSchedule()
			schedule.DayOfWeek = day

			err := schedule.Validate()

			assert.NoError(t, err, "day %d should be valid", day)
		}
	})
}

// =============================================================================
// IsTimeWithinRange Tests
// =============================================================================

func TestOperatingSchedule_IsTimeWithinRange(t *testing.T) {
	t.Run("when time is within range then returns true", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.True(t, result)
	})

	t.Run("when time equals open_time then returns true", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.True(t, result)
	})

	t.Run("when time equals close_time then returns true", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.True(t, result)
	})

	t.Run("when time is before open_time then returns false", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 8, 59, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.False(t, result)
	})

	t.Run("when time is after close_time then returns false", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 18, 1, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.False(t, result)
	})

	t.Run("when open_time is invalid then returns false", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "invalid",
			CloseTime: "18:00",
		}
		checkTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.False(t, result)
	})

	t.Run("when close_time is invalid then returns false", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "invalid",
		}
		checkTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		result := schedule.IsTimeWithinRange(checkTime)

		assert.False(t, result)
	})

	t.Run("ignores the date and only checks time", func(t *testing.T) {
		schedule := &OperatingSchedule{
			OpenTime:  "09:00",
			CloseTime: "18:00",
		}
		// Different dates, same time - should both return true
		time1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		time2 := time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

		assert.True(t, schedule.IsTimeWithinRange(time1))
		assert.True(t, schedule.IsTimeWithinRange(time2))
	})
}

// =============================================================================
// DayOfWeekFromTime Tests
// =============================================================================

func TestDayOfWeekFromTime(t *testing.T) {
	testCases := []struct {
		weekday  time.Weekday
		expected DayOfWeek
	}{
		{time.Sunday, Sunday},
		{time.Monday, Monday},
		{time.Tuesday, Tuesday},
		{time.Wednesday, Wednesday},
		{time.Thursday, Thursday},
		{time.Friday, Friday},
		{time.Saturday, Saturday},
	}

	for _, tc := range testCases {
		t.Run(tc.weekday.String(), func(t *testing.T) {
			result := DayOfWeekFromTime(tc.weekday)

			assert.Equal(t, tc.expected, result)
		})
	}
}
